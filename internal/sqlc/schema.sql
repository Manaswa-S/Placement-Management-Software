CREATE TABLE users (
    user_id      bigint NOT NULL DEFAULT nextval('id_seq'::regclass),
    email        character varying(100) NOT NULL,
    password     character varying(100) NOT NULL,
    role         bigint NOT NULL,
    user_uuid    uuid NOT NULL DEFAULT gen_random_uuid(),
    created_at   timestamp without time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT pk_1 PRIMARY KEY (user_id),
    CONSTRAINT unique_email UNIQUE (email),
    CONSTRAINT unique_uuid UNIQUE (user_uuid)
);

CREATE TABLE companies (
    company_id bigint NOT NULL DEFAULT nextval('companies_company_id_seq'::regclass),
    company_name character varying(100) NOT NULL,
    representative_email character varying(100) NOT NULL,
    representative_contact character varying(10) NOT NULL,
    representative_name character varying(50) NOT NULL,
    data_url text,
    user_id bigint NOT NULL,
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
    active_status boolean NOT NULL DEFAULT true,
    CONSTRAINT jobs_pkey PRIMARY KEY (job_id),
    CONSTRAINT jobs_company_id_fkey FOREIGN KEY (company_id)
        REFERENCES companies(company_id)
        ON DELETE CASCADE
);

CREATE TABLE students (
    student_id BIGINT NOT NULL PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    student_name TEXT NOT NULL,
    roll_number VARCHAR(20) NOT NULL UNIQUE,
    student_dob DATE NOT NULL,
    gender VARCHAR(20) NOT NULL,
    course VARCHAR(20) NOT NULL,
    department TEXT NOT NULL,
    year_of_study VARCHAR(10) NOT NULL,
    resume_url TEXT,
    result_url TEXT NOT NULL UNIQUE,
    cgpa DOUBLE PRECISION,
    contact_no VARCHAR(10) NOT NULL,
    student_email VARCHAR(100) NOT NULL,
    address TEXT,
    skills TEXT,
    user_id BIGINT NOT NULL,
    extras JSON,
    CONSTRAINT uni_result_url UNIQUE (result_url),
    CONSTRAINT uni_roll_no UNIQUE (roll_number),
    CONSTRAINT students_pkey PRIMARY KEY (student_id),
    CONSTRAINT students_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON DELETE CASCADE
);

CREATE TABLE applications (
    application_id BIGINT PRIMARY KEY DEFAULT nextval('applications_application_id_seq'::regclass),
    job_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    data_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status application_status NOT NULL DEFAULT 'applied'::application_status_test,
    CONSTRAINT students_app_pkey FOREIGN KEY (student_id) REFERENCES students(student_id),
    CONSTRAINT jobs_pkey FOREIGN KEY (job_id) REFERENCES jobs(job_id) ON DELETE CASCADE
);

