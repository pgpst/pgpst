package database

import (
	"database/sql"

	"code.pgp.st/pgpst/pkg/config"
)

type migration struct {
	Name    string
	Migrate map[config.Database]func(*sql.DB) error
}

var Migrations = []migration{
	{
		Name: "create_tables",
		Migrate: map[config.Database]func(*sql.DB) error{
			config.Postgres: func(db *sql.DB) error {
				if _, err := db.Exec(`
					CREATE TABLE accounts (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						main_address  text,
						password      bytea,
						susbcription  char(20),
						alt_email     text,
						status        text
					);

					CREATE TABLE addresses (
						id            text primary key,
						styled_id     text,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						public_key    char(16)
					);

					CREATE TABLE applications (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						secret        text,
						callback      text,
						name          text,
						homepage      text,
						description   text
					);

					CREATE TABLE emails (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						message_id    text,
						thread        char(20),
						status        text,
						public_key    char(16),
						manifest      bytea,
						body          text
					);

					CREATE TABLE public_keys (
						id            char(16) primary key,
						fingerprint   char(40),
						id_short      char(8),
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						body          text,
						algorithm     int,
						length        int,
						identities    json
					);

					CREATE TABLE labels (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						name          text,
						is_system     bool
					);

					CREATE TABLE resources (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						body          text,
						tags          text[],
						meta          json
					);

					CREATE TABLE threads (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						address       text,
						members       bytea[],
						is_read       bool,
						last_read     char(20),
						secure        text
					);

					CREATE TABLE thread_labels (
						thread_id     char(20),
						label_id      char(20)
					);

					CREATE TABLE tokens (
						id            char(20) primary key,
						date_created  timestamp with time zone,
						date_modified timestamp with time zone,
						owner         char(20),
						expiry_date   timestamp with time zone,
						type          string,
						scope         text[],
						app_id        char(20)
					);

					ALTER TABLE accounts ADD CONSTRAINT accounts_main_address_fkey
						FOREIGN KEY (main_address) REFERENCES addresses (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE RESTRICT;

					ALTER TABLE addresses ADD CONSTRAINT addresses_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE addresses ADD CONSTRAINT addresses_public_key_fkey
						FOREIGN KEY (public_key) REFERENCES public_keys (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE RESTRICT;

					ALTER TABLE applications ADD CONSTRAINT applications_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE emails ADD CONSTRAINT emails_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE emails ADD CONSTRAINT emails_public_key_fkey
						FOREIGN KEY (public_key) REFERENCES public_keys (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE RESTRICT;

					ALTER TABLE emails ADD CONSTRAINT emails_thread_fkey
						FOREIGN KEY (thread) REFERENCES threads (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE labels ADD CONSTRAINT labels_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE public_keys ADD CONSTRAINT public_keys_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE resources ADD CONSTRAINT resources_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE threads ADD CONSTRAINT threads_address_fkey
						FOREIGN KEY (address) REFERENCES addresses (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE RESTRICT;

					ALTER TABLE threads ADD CONSTRAINT threads_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE threads ADD CONSTRAINT threads_last_read_fkey
						FOREIGN_KEY (last_read) REFERENCES emails (id)
						MATCH SIMPLE ON UPDATE CASCADE;

					ALTER TABLE threads ADD CONSTRAINT thread_labels_thread_fkey
						FOREIGN_KEY (thread) REFERENCES threads (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE threads ADD CONSTRAINT thread_labels_label_fkey
						FOREIGN_KEY (label) REFERENCES labels (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE RESTRICT;

					ALTER TABLE tokens ADD CONSTRAINT tokens_owner_fkey
						FOREIGN KEY (owner) REFERENCES accounts (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;

					ALTER TABLE tokens ADD CONSTRAINT tokens_app_id_fkey
						FOREIGN KEY (app_id) REFERENCES applications (id)
						MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE;
				`); err != nil {
					return err
				}
				return nil
			},
			config.SQLite: func(db *sql.DB) error {
				if _, err := db.Exec(`
					CREATE TABLE accounts (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						/* main_address  text, */
						password      blob,
						susbcription  character(20),
						alt_email     text,
						status        text
					);

					CREATE TABLE public_keys (
						id            character(16) primary key,
						fingerprint   character(40),
						id_short      character(8),
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						body          text,
						algorithm     int,
						length        int,
						identities    json
					);

					CREATE TABLE addresses (
						id            text primary key,
						styled_id     text,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						public_key    character(16)
							REFERENCES public_keys(id) ON UPDATE CASCADE ON DELETE RESTRICT
					);

					CREATE TABLE applications (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						secret        text,
						callback      text,
						name          text,
						homepage      text,
						description   text
					);

					CREATE TABLE threads (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						address       text
							REFERENCES addresses(id) ON UPDATE CASCADE ON DELETE RESTRICT,
						members       bytea[],
						is_read       bool,
						/* last_read     character(20), */
						secure        text
					);

					CREATE TABLE emails (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						message_id    text,
						thread        character(20)
							REFERENCES threads(id) ON UPDATE CASCADE ON DELETE CASCADE,
						status        text,
						public_key    character(16)
							REFERENCES public_keys(id) ON UPDATE CASCADE ON DELETE RESTRICT,
						manifest      bytea,
						body          text
					);

					CREATE TABLE labels (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						name          text,
						is_system     bool
					);

					CREATE TABLE resources (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						body          text,
						tags          text[],
						meta          json
					);

					CREATE TABLE thread_labels (
						thread_id     character(20)
							REFERENCES threads(id) ON UPDATE CASCADE ON DELETE CASCADE,
						label_id      character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE RESTRICT
					);

					CREATE TABLE tokens (
						id            character(20) primary key,
						date_created  datetime,
						date_modified datetime,
						owner         character(20)
							REFERENCES accounts(id) ON UPDATE CASCADE ON DELETE CASCADE,
						expiry_date   datetime,
						type          string,
						scope         text[],
						app_id        character(20)
							REFERENCES applications(id) ON UPDATE CASCADE ON DELETE CASCADE
					);

					ALTER TABLE accounts ADD COLUMN main_address text
						REFERENCES addresses(id) ON UPDATE CASCADE ON DELETE RESTRICT;

					ALTER TABLE threads ADD COLUMN last_read character(20)
						REFERENCES emails(id) ON UPDATE CASCADE;
				`); err != nil {
					return err
				}

				return nil
			},
		},
	},
}
