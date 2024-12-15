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

CREATE TABLE companies (
    company_id bigint NOT NULL DEFAULT nextval('companies_company_id_seq'::regclass),
    company_name character varying(100) NOT NULL,
    company_email character varying(100) NOT NULL,
    representative_contact character varying(10) NOT NULL,
    representative_name character varying(50) NOT NULL,
    data_url text,
    user_id bigint NOT NULL,
    is_verified boolean DEFAULT false,
    CONSTRAINT companies_pkey PRIMARY KEY (company_id),
    CONSTRAINT uni_comp_name UNIQUE (company_name),
    CONSTRAINT uni_email UNIQUE (email),
    CONSTRAINT companies_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
);

CREATE TABLE jobs (
    job_id bigint NOT NULL DEFAULT nextval('jobs_job_id_seq'::regclass),
    data_url text,
    created_at timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    company_id bigint NOT NULL,
    title text NOT NULL,
    location text NOT NULL,
    type text NOT NULL,
    salary text NOT NULL,
    skills text[] NOT NULL,
    position text NOT NULL,
    extras JSON,
    CONSTRAINT jobs_pkey PRIMARY KEY (job_id),
    CONSTRAINT jobs_company_id_fkey FOREIGN KEY (company_id)
        REFERENCES companies(company_id)
        ON DELETE CASCADE
);
