CREATE TABLE theredshirts_lobby.lobby (
    id uuid PRIMARY KEY NOT NULL,
    status varchar NOT NULL,
    name varchar NOT NULL,
    owner uuid NOT NULL,
    password varchar NOT NULL,
    difficulty number NOT NULL,
    mission_length number NOT NULL,
    crew_members number NOT NULL,
    max_players number NOT NULL,
    expansion_packs varchar[] 
);
