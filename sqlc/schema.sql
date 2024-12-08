CREATE TABLE "public".users
(
 user_id    bigint NOT NULL,
 first_name varchar(50) NOT NULL,
 last_name  varchar(50) NULL,
 email      varchar(100) NOT NULL,
 password   varchar(100) NOT NULL,
 role       bigint NOT NULL,
 user_uuid  uuid NOT NULL,
 created_at timestamp with time zone NOT NULL,
 CONSTRAINT PK_1 PRIMARY KEY ( user_id )
);

