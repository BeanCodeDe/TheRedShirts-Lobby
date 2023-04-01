package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

const (
	player_table_name              = "player"
	create_player_sql              = "INSERT INTO %s.%s(id, name, lobby_id, last_refresh, payload) VALUES($1, $2, $3, $4, $5)"
	update_player_sql              = "UPDATE %s.%s SET name = $2, lobby_id = $3, last_refresh = $4, payload = $5  WHERE id = $1"
	delete_player_sql              = "DELETE FROM %s.%s WHERE id = $1"
	delete_player_in_lobby_sql     = "DELETE FROM %s.%s WHERE lobby_id = $1"
	delete_player_older_then_sql   = "DELETE FROM %s.%s WHERE last_refresh < $1"
	select_player_by_player_id_sql = "SELECT id, name, lobby_id, payload FROM %s.%s WHERE id = $1"
	select_player_by_lobby_id_sql  = "SELECT id, name, lobby_id, payload FROM %s.%s WHERE lobby_id = $1"
)

var (
	ErrPlayerAlreadyExists = errors.New("player already exists")
)

func (tx *postgresTransaction) CreatePlayer(player *Player) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(create_player_sql, schema_name, player_table_name), player.ID, player.Name, player.LobbyId, player.LastRefresh, player.Payload); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return ErrPlayerAlreadyExists
			}
		}

		return fmt.Errorf("unknown error when inserting player: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) UpdatePlayer(player *Player) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(update_player_sql, schema_name, player_table_name), player.ID, player.Name, player.LobbyId, player.LastRefresh, player.Payload); err != nil {
		return fmt.Errorf("unknown error when updating player: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) DeletePlayer(id uuid.UUID) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(delete_player_sql, schema_name, player_table_name), id); err != nil {
		return fmt.Errorf("unknown error when deliting player: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) DeletePlayerOlderRefreshDate(time time.Time) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(delete_player_older_then_sql, schema_name, player_table_name), time); err != nil {
		return fmt.Errorf("unknown error when deliting players older refreshe date: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) DeleteAllPlayerInLobby(lobbyId uuid.UUID) error {
	if _, err := tx.tx.Exec(context.Background(), fmt.Sprintf(delete_player_in_lobby_sql, schema_name, player_table_name), lobbyId); err != nil {
		return fmt.Errorf("unknown error when deliting players from lobby: %v", err)
	}
	return nil
}

func (tx *postgresTransaction) GetPlayerById(id uuid.UUID) (*Player, error) {
	var players []*Player
	if err := pgxscan.Select(context.Background(), tx.tx, &players, fmt.Sprintf(select_player_by_player_id_sql, schema_name, player_table_name), id); err != nil {
		return nil, fmt.Errorf("error while selecting player with id %v: %v", id, err)
	}

	if len(players) == 0 {
		return nil, nil
	}

	if len(players) != 1 {
		return nil, fmt.Errorf("cant find only one player. Players: %v", players)
	}

	return players[0], nil
}

func (tx *postgresTransaction) GetAllPlayersInLobby(lobbyId uuid.UUID) ([]*Player, error) {
	var players []*Player
	if err := pgxscan.Select(context.Background(), tx.tx, &players, fmt.Sprintf(select_player_by_lobby_id_sql, schema_name, player_table_name), lobbyId); err != nil {
		return nil, fmt.Errorf("error while selecting all players in lobby: %v", err)
	}

	return players, nil
}
