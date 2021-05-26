package dbrepo

import (
	"errors"
	"time"

	"github.com/andkolbe/bookings/internal/models"
)


func (m *testDBRepo) AllUsers() bool {
	return true
}

// inserts a reservation into the db
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if the room id is 2, then fail; otherwise, pass
	if res.RoomID == 2 {
		return 0, errors.New("some error")
	}
	return 1, nil
}

// inserts a room restriction into the db
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	// when trying to insert a room restriction for room id 1000, fail
	if r.RoomID == 1000 {
		return errors.New("some error")
	}
	return nil
}

// returns true if availability exists for roomID
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// returns a slice of available rooms, if any, for given date range
// we don't need the room id as a param because we are searching through all rooms
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

// gets a room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 { // if the room id is greater than 2, throw an error
		return room, errors.New("Some error")
	}
	return room, nil
}

// gets a room by id
func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User
	
	return u, nil
}

func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {  
	return 1, "", nil
}