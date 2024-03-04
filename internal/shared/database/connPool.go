/*******************************************************************************
 * Copyright 2018-2024 Aaron Hnatiw
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 ******************************************************************************/

package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pgxuuid "github.com/jackc/pgx-gofrs-uuid"
)

// GetDbConnPool parses the given config byte array (created from TOML config file) and returns a database
// connection pool.
func GetDbConnPool(dbConnString string) (*pgxpool.Pool, error) {
	// Create a database connection config
	pgxPoolConfig, configErr := pgxpool.ParseConfig(dbConnString)
	if configErr != nil {
		return nil, fmt.Errorf("unable to parse database connection string: %w", configErr)
	}

	// Add support for UUIDs
	pgxPoolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		pgxuuid.Register(conn.TypeMap())
		return nil
	}

	// Create the database connection pool
	dbConnPool, connErr := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if connErr != nil {
		return nil, fmt.Errorf("unable to create database connection pool from database config: %w", connErr)
	}

	return dbConnPool, nil
}
