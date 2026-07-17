package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"vitalis/internal/enviroment"

	_ "github.com/lib/pq"
)

var Database *sql.DB

func Connect() error {
	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			enviroment.ENV.DATABASE_ADDRESS,
			enviroment.ENV.DATABASE_PORT,
			enviroment.ENV.DATABASE_NAME,
			enviroment.ENV.DATABASE_USER,
			enviroment.ENV.DATABASE_PASSWORD))

	if err != nil {
		log.Fatal(err)
		return err
	}

	Database = db

	return nil
}

func Initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS public.auth_session_keys (
			id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
			session_key varchar(512) NOT NULL,
			creation_datetime timestamp DEFAULT now() NOT NULL,
			valid_until timestamp NOT NULL,

			CONSTRAINT auth_session_keys_pk PRIMARY KEY (id)
		);`,

		`CREATE TABLE IF NOT EXISTS public.info_cpu (
			id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
			group_id int4 NOT NULL,
			name varchar(512) NOT NULL,
			physical_cores int4 NOT NULL,
			logical_cores int4 NOT NULL,
			utilization float8 NOT NULL,
			current_speed_mhz float8 NOT NULL,
			base_speed_mhz float8 NOT NULL,
			processes_amount int4 NOT NULL,
			threads_amount int4 NOT NULL,
			handles_amount int NOT NULL,
			uptime int8 NOT NULL,
			insertion_datetime timestamp DEFAULT now() NOT NULL,

			CONSTRAINT info_cpu_pk PRIMARY KEY (id)
		);`,

		`CREATE SEQUENCE IF NOT EXISTS public.info_cpu_group_id_seq
			AS int4
			START WITH 1
			INCREMENT BY 1
			NO MINVALUE
			NO MAXVALUE
			CACHE 1;`,

		`CREATE TABLE IF NOT EXISTS public.info_ram (
			id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
			group_id int4 NOT NULL,
			total int8 NOT NULL,
			used int8 NOT NULL,
			free int8 NOT NULL,
			commited int8 NOT NULL,
			cached int8 NOT NULL,
			insertion_datetime timestamp DEFAULT now() NOT NULL,

			CONSTRAINT info_ram_pk PRIMARY KEY (id)
		);`,

		`CREATE SEQUENCE IF NOT EXISTS public.info_ram_group_id_seq
			AS int4
			START WITH 1
			INCREMENT BY 1
			NO MINVALUE
			NO MAXVALUE
			CACHE 1;`,

		`CREATE TABLE IF NOT EXISTS public.info_net (
			id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
			group_id int4 NOT NULL,
			bytes_sent int8 NOT NULL,
			bytes_recv int8 NOT NULL,
			packets_sent int8 NOT NULL,
			packets_recv int8 NOT NULL,
			err_in int8 NOT NULL,
			err_out int8 NOT NULL,
			connections int4 NOT NULL,
			insertion_datetime timestamp DEFAULT now() NOT NULL,

			CONSTRAINT info_net_pk PRIMARY KEY (id)
		);`,

		`CREATE SEQUENCE IF NOT EXISTS public.info_net_group_id_seq
			AS int4
			START WITH 1
			INCREMENT BY 1
			NO MINVALUE
			NO MAXVALUE
			CACHE 1;`,

		`CREATE TABLE IF NOT EXISTS public.info_file (
			id int4 GENERATED ALWAYS AS IDENTITY NOT NULL,
			group_id int4 NOT NULL,
			path text NOT NULL,
			total int8 NOT NULL,
			used int8 NOT NULL,
			free int8 NOT NULL,
			used_percent float8 NOT NULL,
			insertion_datetime timestamp DEFAULT now() NOT NULL,
			
			CONSTRAINT info_file_pk PRIMARY KEY (id)
		);`,

		`CREATE SEQUENCE IF NOT EXISTS public.info_file_group_id_seq
			AS int4
			START WITH 1
			INCREMENT BY 1
			NO MINVALUE
			NO MAXVALUE
			CACHE 1;`,
	}

	for _, query := range queries {
		_, err := Database.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func Validate() {
	err := Database.Ping()

	if err != nil {
		fmt.Printf("Validation failed. Re-connecting to the database...")
	}

	err = Connect()
	if err != nil {
		log.Fatalf("Failed to reconnect to PostgreSQL: %v, revalidating...", err)
		Validate()
	}
}

func GetDatabaseTime() (*time.Time, error) {
	rows, err := Database.Query(`select now();`)

	if err != nil {
		log.Printf("Error occured in GetDatabaseTime(): %v", err)
		return nil, err
	}

	var currentTimestamp time.Time

	for rows.Next() {
		if err = rows.Scan(&currentTimestamp); err != nil {
			log.Printf("Error occured in GetDatabaseTime(): %v", err)
			return nil, err
		}
	}

	return &currentTimestamp, nil
}
