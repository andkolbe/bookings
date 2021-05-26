package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/andkolbe/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// inserts a reservation into the db
func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if a user loses their connection to the internet while in the middle of submitting data to the db, we want that to close and not go through
	// this is called a transaction
	// Go uses something called context to fix this
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	var newID int

	stmt := `INSERT INTO reservations (first_name, last_name, email, phone, start_date, end_date, room_id,
				created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// inserts a room restriction into the db
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	stmt := `INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)
	if err != nil {
		return err
	}

	return nil
}

// returns true if availability exists for roomID
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	var numRows int

	// make sure the dates for the room are not already taken
	query := `
		SELECT COUNT(id)
		FROM room_restrictions
		WHERE room_id = $1 AND $2 < end_date AND $3 > start_date
	`
	// run the query in PG
	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	// scan the value in numRows
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		// return the numRows. If it returns true, the room is available
		return true, nil
	}
	return false, nil

}

// returns a slice of available rooms, if any, for given date range
// we don't need the room id as a param because we are searching through all rooms
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	var rooms []models.Room

	query := `
		SELECT rooms.id, rooms.room_name
		FROM rooms
		WHERE rooms.id NOT IN (
			SELECT room_id
			FROM room_restrictions
			WHERE $1 < room_restrictions.end_date AND $2 > room_restrictions.start_date
			)
	`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil

}

// gets a room by id
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	var room models.Room

	query := `
		SELECT id, room_name, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room, err
	}

	return room, nil
}

// returns a user by ID
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	query := `SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at FROM users WHERE id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

// updates a user in the db
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	query := `UPDATE users SET first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// authenticates a user
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // cancel transaction if it takes longer than 3 seconds to complete
	defer cancel()

	// declare two variables to store info we get from the db
	var id int 
	var hashedPassword string

	// query the db
	row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = $1", email)
	// store id and password values in id and hashedPassword variables
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// compare the hash we grabbed from the db against the hash created from the password that the user typed in the form
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	// if they don't match
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
		// if there's another error that isn't a mismatched password
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}
