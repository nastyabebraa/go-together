package entity

import (
	"database/sql"
)

type Event struct {
	IDEvent          int64        `json:"id_event"`
	DateOfTrip       sql.NullTime `json:"date_of_trip"`
	AvailableSeats   int          `json:"available_seats"`
	TripCost         float64      `json:"trip_cost"`
	CostPerPerson    float64      `json:"cost_per_person"`
	DepartureAddress string       `json:"departure_address"`
	ArrivalAddress   string       `json:"arrival_address"`
	CarNumber        string       `json:"car_number"`
	UserID           int64        `json:"user_id"`
	DriverName       string       `json:"driver_name"`
	Status           bool         `json:"status"`
}
