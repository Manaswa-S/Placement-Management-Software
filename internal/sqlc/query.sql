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



-- name: InsertNewApplication :exec
INSERT INTO applications (job_id, student_id, data_url) 
VALUES ($1, (SELECT student_id FROM students WHERE user_id = $2), $3);



-- name: GetApplicableJobs :many
SELECT 
    jobs.job_id,
    jobs.title, 
    jobs.location,
    jobs.type,
    jobs.salary,
    jobs.position,
    jobs.skills,
    jobs.company_id,
    jobs.active_status,
    companies.company_name 
FROM jobs
JOIN companies ON jobs.company_id = companies.company_id 
LEFT JOIN (SELECT applications.job_id FROM applications WHERE applications.student_id = (SELECT student_id FROM students WHERE students.user_id = $1)) AS t 
ON jobs.job_id = t.job_id
WHERE t.job_id IS NULL;

-- name: GetMyApplications :many
SELECT 
    jobs.job_id,
    jobs.title, 
    jobs.location,
    jobs.type,
    jobs.salary,
    jobs.position,
    jobs.skills,
    jobs.company_id,
    companies.company_name,
    companies.representative_email,
    companies.representative_name,
    t.status::TEXT AS status
FROM jobs
JOIN (SELECT applications.job_id, applications.status FROM applications WHERE student_id = (
    SELECT student_id FROM students WHERE students.user_id = $1)) AS t 
ON jobs.job_id = t.job_id
JOIN companies ON jobs.company_id = companies.company_id;



-- name: CancelApplication :exec
DELETE FROM applications 
WHERE student_id = (SELECT student_id FROM students WHERE students.user_id = $1) 
AND job_id = $2;



-- name: GetApplicants :many
SELECT
    students.student_id,
    students.student_name,
    students.roll_number,
    students.gender,
    students.department,
    students.student_email,
    students.contact_no,
    students.cgpa,
    students.skills,
    jobs.job_id, 
    jobs.title, 
    applications.status::TEXT AS status 
FROM applications
JOIN jobs ON applications.job_id = jobs.job_id
JOIN students ON applications.student_id = students.student_id
WHERE jobs.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1);


-- name: GetResumePath :one
SELECT resume_url, result_url FROM students WHERE student_id = $1;


-- name: GetJobListings :many
SELECT 
    jobs.job_id,
    jobs.created_at::DATE as created_at,
    jobs.title,
    jobs.location,
    jobs.type,
    jobs.salary,
    jobs.skills,
    jobs.position,
    jobs.active_status,
    COALESCE(t.no_of_applications, 0)    
FROM jobs
LEFT JOIN (SELECT job_id, COUNT(job_id) AS no_of_applications FROM applications GROUP BY job_id) AS t
ON jobs.job_id = t.job_id
WHERE jobs.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1);


-- name: CloseJob :exec
UPDATE jobs
SET active_status = false
WHERE jobs.job_id = $1 
AND jobs.company_id = (SELECT companies.company_id FROM companies WHERE user_id = $2);

-- name: DeleteJob :exec
DELETE FROM jobs 
WHERE jobs.job_id = $1
AND jobs.company_id = (SELECT companies.company_id FROM companies WHERE user_id = $2);