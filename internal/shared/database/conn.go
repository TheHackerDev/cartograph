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
)

// GetDbConn parses the given config byte array (created from TOML config file) and returns a database connection.
func GetDbConn(dbConnString string) (*pgx.Conn, error) {
	// Create a database connection config
	pgxConfig, configErr := pgx.ParseConfig(dbConnString)
	if configErr != nil {
		return nil, fmt.Errorf("unable to parse database connection string: %w", configErr)
	}

	// Create the database connection pool
	dbConn, connErr := pgx.ConnectConfig(context.Background(), pgxConfig)
	if connErr != nil {
		return nil, fmt.Errorf("unable to create database connection from database config: %w", connErr)
	}

	return dbConn, nil
}
