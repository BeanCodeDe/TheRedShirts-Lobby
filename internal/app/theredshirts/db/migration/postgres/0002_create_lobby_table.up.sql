CREATE TABLE theredshirts_lobby.lobby (
    id uuid PRIMARY KEY NOT NULL,
    name varchar NOT NULL,
    owner uuid NOT NULL,
    password varchar NOT NULL,
    difficulty varchar NOT NULL
);
