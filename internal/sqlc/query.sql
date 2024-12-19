-- name: GetAll :many
SELECT * FROM users;

-- name: SignupUser :one
INSERT INTO users (email, password, role) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserData :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateEmailConfirmation :exec
UPDATE users
SET confirmed = true
WHERE email = $1;

-- name: UpdatePassword :exec
UPDATE users
SET password = $2
WHERE email = $1;



-- >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
-- Company queries 

-- name: InsertNewJob :one
INSERT INTO jobs (data_url, company_id, title, location, type, salary, skills, position, extras)
VALUES ($1, (SELECT company_id FROM companies WHERE representative_email = $2), $3, $4, $5, $6, $7, $8, $9)
RETURNING *;


-- name: ExtraInfoCompany :one
INSERT INTO companies (company_name, representative_email, representative_contact, representative_name, data_url, user_id)
VALUES ($1, $2, $3, $4, $5, (SELECT user_id FROM users WHERE email = $6))
RETURNING *;


-- name: ExtraInfoStudent :one
INSERT INTO students (student_name, roll_number, student_dob, gender, course, department, year_of_study, resume_url, result_url, cgpa, contact_no, student_email, address, skills, user_id, extras)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, (SELECT user_id FROM users WHERE email = $15), $16)
RETURNING *;

