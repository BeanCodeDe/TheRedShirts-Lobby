CREATE INDEX player_lobby_idx ON theredshirts_lobby.player (lobby_id);
CREATE INDEX player_lobby_without_spectator_idx ON theredshirts_lobby.player (lobby_id) WHERE spectator = false;
CREATE INDEX player_refresh_idx ON theredshirts_lobby.player (last_refresh desc);