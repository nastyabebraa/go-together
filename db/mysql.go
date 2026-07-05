package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"ride-together-bot/entity"
)

type DB struct {
	Conn *sql.DB
}

func NewDatabase(ctx context.Context, dsn string) (*DB, error) {
	connection, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := connection.PingContext(ctx); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &DB{Conn: connection}, nil
}

func (db *DB) Close() error {
	if db == nil || db.Conn == nil {
		return nil
	}
	return db.Conn.Close()
}

func (db *DB) GetAllDataFromEvents(ctx context.Context, departureAddress string) ([]entity.Event, error) {
	const query = `
		SELECT id_events, DATE_FORMAT(date_of_trip, '%Y-%m-%d'),
		       COALESCE(available_seats, 0), COALESCE(trip_cost, 0),
		       COALESCE(cost_per_person, 0), COALESCE(departure_address, ''),
		       COALESCE(arrival_address, ''), COALESCE(car_number, ''),
		       COALESCE(user_id, 0), COALESCE(driver_name, ''), COALESCE(status, false)
		FROM events
		WHERE departure_address = ? AND status = true`

	rows, err := db.Conn.QueryContext(ctx, query, departureAddress)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	events := make([]entity.Event, 0)
	for rows.Next() {
		var event entity.Event
		var date sql.NullString
		if err := rows.Scan(
			&event.IDEvent,
			&date,
			&event.AvailableSeats,
			&event.TripCost,
			&event.CostPerPerson,
			&event.DepartureAddress,
			&event.ArrivalAddress,
			&event.CarNumber,
			&event.UserID,
			&event.DriverName,
			&event.Status,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		if date.Valid {
			parsed, err := time.Parse(time.DateOnly, date.String)
			if err != nil {
				return nil, fmt.Errorf("parse trip date: %w", err)
			}
			event.DateOfTrip = sql.NullTime{Time: parsed, Valid: true}
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return events, nil
}

func (db *DB) RegisterUser(ctx context.Context, user entity.User) error {
	user.Name = strings.TrimSpace(user.Name)
	user.Phone = strings.TrimSpace(user.Phone)
	user.Login = strings.TrimSpace(user.Login)
	if user.Name == "" || user.Phone == "" || user.ChatID == 0 {
		return errors.New("user data is incomplete")
	}

	const query = `
		INSERT INTO users (name, phone, login, chatID)
		VALUES (?, ?, NULLIF(?, ''), ?)`
	_, err := db.Conn.ExecContext(
		ctx,
		query,
		user.Name,
		user.Phone,
		user.Login,
		user.ChatID,
	)
	if err != nil {
		return fmt.Errorf("register user: %w", err)
	}
	return nil
}

func (db *DB) IsExists(ctx context.Context, chatID int64) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE chatID = ?)`
	var exists bool
	if err := db.Conn.QueryRowContext(ctx, query, chatID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user existence: %w", err)
	}
	return exists, nil
}

func (db *DB) UpsertLocation(ctx context.Context, chatID int64, latitude, longitude float64) error {
	userID, err := db.GetUserID(ctx, chatID)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO user_location (user_id, latitude, longitude, created_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			latitude = VALUES(latitude),
			longitude = VALUES(longitude),
			created_at = VALUES(created_at)`
	_, err = db.Conn.ExecContext(ctx, query, userID, latitude, longitude, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("upsert location: %w", err)
	}
	return nil
}

func (db *DB) GetUserID(ctx context.Context, chatID int64) (int64, error) {
	const query = `SELECT id FROM users WHERE chatID = ?`
	var id int64
	if err := db.Conn.QueryRowContext(ctx, query, chatID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, entity.ErrUserNotFound
		}
		return 0, fmt.Errorf("get user ID: %w", err)
	}
	return id, nil
}

func (db *DB) IsDriver(ctx context.Context, chatID int64) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM events
			JOIN users ON users.id = events.user_id
			WHERE users.chatID = ?
		)`
	var exists bool
	if err := db.Conn.QueryRowContext(ctx, query, chatID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check driver status: %w", err)
	}
	return exists, nil
}

func (db *DB) GetAllDepartureAddresses(ctx context.Context) ([]string, error) {
	const query = `SELECT DISTINCT departure_address FROM events WHERE status = true`
	rows, err := db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query departure addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]string, 0)
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, fmt.Errorf("scan departure address: %w", err)
		}
		addresses = append(addresses, address)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate departure addresses: %w", err)
	}
	return addresses, nil
}

func (db *DB) GetEventIDs(ctx context.Context, chatID int64) ([]int64, error) {
	const query = `
		SELECT DISTINCT event_id
		FROM passengers
		WHERE user_chatID = ?
		UNION
		SELECT events.id_events
		FROM events
		JOIN users ON users.id = events.user_id
		WHERE users.chatID = ?
		ORDER BY 1`
	rows, err := db.Conn.QueryContext(ctx, query, chatID, chatID)
	if err != nil {
		return nil, fmt.Errorf("query event IDs: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan event ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event IDs: %w", err)
	}
	return ids, nil
}
