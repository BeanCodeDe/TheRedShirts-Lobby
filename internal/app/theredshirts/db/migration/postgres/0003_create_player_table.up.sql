CREATE TABLE theredshirts_lobby.player (
    id uuid PRIMARY KEY NOT NULL,
    name varchar NOT NULL,
    lobby_id uuid NOT NULL REFERENCES theredshirts_lobby.lobby(id),
    last_refresh timestamp NOT NULL,
    payload json
);
