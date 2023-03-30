package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

const (
	lobby_table_name       = "lobby"
	create_lobby_sql       = "INSERT INTO %s.%s(id, status, name, owner, password, difficulty, mission_length, crew_members, max_players, expansion_packs) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	update_lobby_sql       = "UPDATE %s.%s SET status = $2, name = $3, owner = $4, password = $5, difficulty = $6, mission_length = $7, crew_members = $8, max_players = $9, expansion_packs = $10 WHERE id = $1"
	delete_lobby_sql       = "DELETE FROM %s.%s WHERE id = $1"
	delete_empty_lobby_sql = "DELETE FROM %s.%s WHERE NOT EXISTS (SELECT 1 FROM %s.%s AS p WHERE p.lobby_id = lobby.id)"
	select_lobby_by_id_sql = "SELECT id, status, name, owner, password, difficulty, mission_length, crew_members, max_players, expansion_packs FROM %s.%s WHERE id = $1"
	select_lobby_sql       = "SELECT id, status, name, owner, password, difficulty, mission_length, crew_members, max_players, expansion_packs FROM %s.%s"
)

var (
	ErrLobbyAlreadyExists = errors.New("lobby already exists")
)

func (tx *postgresTransaction) CreateLobby(lobby *Lobby) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(create_lobby_sql, schema_name, lobby_table_name), lobby.ID, lobby.Status, lobby.Name, lobby.Owner, lobby.Password, lobby.Difficulty, lobby.MissionLength, lobby.CrewMembers, lobby.MaxPlayers, lobby.ExpansionPacks); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return ErrLobbyAlreadyExists
			}
		}

		return fmt.Errorf("unknown error when inserting lobby: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) UpdateLobby(lobby *Lobby) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(update_lobby_sql, schema_name, lobby_table_name), lobby.ID, lobby.Status, lobby.Name, lobby.Owner, lobby.Password, lobby.Difficulty, lobby.MissionLength, lobby.CrewMembers, lobby.MaxPlayers, lobby.ExpansionPacks); err != nil {
		return fmt.Errorf("unknown error when updating lobby: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) DeleteLobby(id uuid.UUID) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(delete_lobby_sql, schema_name, lobby_table_name), id); err != nil {
		return fmt.Errorf("unknown error when deliting lobby: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) DeleteEmptyLobbies() error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(delete_empty_lobby_sql, schema_name, lobby_table_name, schema_name, player_table_name)); err != nil {
		return fmt.Errorf("unknown error when deliting lobby: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) GetLobbyById(id uuid.UUID) (*Lobby, error) {
	var lobbies []*Lobby
	if err := pgxscan.Select(context.Background(), tx.tx, &lobbies, fmt.Sprintf(select_lobby_by_id_sql, schema_name, lobby_table_name), id); err != nil {
		return nil, fmt.Errorf("error while selecting lobby with id %v: %v", id, err)
	}

	if len(lobbies) == 0 {
		return nil, nil
	}

	if len(lobbies) != 1 {
		return nil, fmt.Errorf("cant find only one lobby. Lobbies: %v", lobbies)
	}

	return lobbies[0], nil
}

func (tx *postgresTransaction) GetAllLobbies() ([]*Lobby, error) {
	var lobbies []*Lobby
	if err := pgxscan.Select(context.Background(), tx.tx, &lobbies, fmt.Sprintf(select_lobby_sql, schema_name, lobby_table_name)); err != nil {
		return nil, fmt.Errorf("error while selecting all lobbies: %v", err)
	}

	return lobbies, nil
}
