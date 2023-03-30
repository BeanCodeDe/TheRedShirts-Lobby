CREATE TABLE theredshirts_lobby.lobby (
    id uuid PRIMARY KEY NOT NULL,
    status varchar NOT NULL,
    name varchar NOT NULL,
    owner uuid NOT NULL,
    password varchar NOT NULL,
    difficulty integer NOT NULL,
    mission_length integer NOT NULL,
    crew_members integer NOT NULL,
    max_players integer NOT NULL,
    expansion_packs varchar[] 
);
