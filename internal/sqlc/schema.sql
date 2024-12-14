CREATE TABLE users (
    user_id      bigint NOT NULL DEFAULT nextval('id_seq'::regclass),
    email        character varying(100) NOT NULL,
    password     character varying(100) NOT NULL,
    role         bigint NOT NULL,
    user_uuid    uuid NOT NULL DEFAULT gen_random_uuid(),
    created_at   timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT pk_1 PRIMARY KEY (user_id),
    CONSTRAINT unique_email UNIQUE (email),
    CONSTRAINT unique_uuid UNIQUE (user_uuid)
);
