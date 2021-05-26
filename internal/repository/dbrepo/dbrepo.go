package dbrepo

import (
	"database/sql"

	"github.com/andkolbe/bookings/internal/config"
	"github.com/andkolbe/bookings/internal/repository"
)


type postgresDBRepo struct {
	App *config.AppConfig
	DB *sql.DB
}

// we pass this function our connection pool and app config, and return a repository
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB: conn,
	}
}