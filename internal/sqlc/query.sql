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


-- name: GetApplicableJobsTypeFilter :many
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
WHERE t.job_id IS NULL 
AND (jobs.type = $2 OR $2 = 'All');




-- name: GetMyApplicationsStatusFilter :many
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
    applications.status::TEXT AS status
FROM applications
JOIN students ON applications.student_id = students.student_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE students.user_id = $1 
  AND ($2 = 'All' OR applications.status::TEXT = $2)
ORDER BY jobs.job_id;



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
    applications.status::TEXT AS status,
    COALESCE(interviews.status::TEXT, '') AS interview_status,
    applications.application_id
FROM applications
JOIN jobs ON applications.job_id = jobs.job_id
JOIN students ON applications.student_id = students.student_id
LEFT JOIN interviews ON applications.application_id = interviews.application_id
WHERE jobs.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1)
AND (jobs.job_id = $2 OR $2 = 0)
AND (applications.status != 'Rejected')
ORDER BY jobs.job_id;


-- name: GetJobDetails :one
SELECT 
    jobs.title,
    companies.company_name
FROM jobs 
JOIN companies ON jobs.company_id = companies.company_id
WHERE jobs.job_id = $1;

-- name: GetAllApplicantsEmailsForJob :many
SELECT 
    students.student_email
FROM applications
JOIN students ON applications.student_id = students.student_id
WHERE applications.job_id = $1;

-- name: GetResumeAndResultPath :one
SELECT 
    resume_url, 
    result_url
FROM students 
JOIN applications 
ON applications.student_id = students.student_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE applications.application_id = $1
AND companies.user_id = $2;


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
LEFT JOIN (
    SELECT 
        job_id, 
        COUNT(job_id) AS no_of_applications 
    FROM applications 
    WHERE status != 'Rejected' 
    GROUP BY job_id ) AS t
ON jobs.job_id = t.job_id
WHERE jobs.company_id = (
    SELECT 
        companies.company_id 
    FROM companies 
    WHERE companies.user_id = $1);


-- name: CloseJob :exec
UPDATE jobs
SET active_status = false
WHERE jobs.job_id = $1 
AND jobs.company_id = (SELECT companies.company_id FROM companies WHERE user_id = $2);

-- name: DeleteJob :exec
DELETE FROM jobs 
WHERE jobs.job_id = $1
AND jobs.company_id = (SELECT companies.company_id FROM companies WHERE user_id = $2);



-- name: GetUserIDCompanyIDJobIDApplicationID :one
SELECT 
    companies.user_id
FROM companies
JOIN jobs ON jobs.company_id = companies.company_id
JOIN applications ON applications.job_id = jobs.job_id
WHERE applications.application_id = $1;

-- name: ApplicationStatusToAnd :exec
UPDATE applications
SET status = $1
WHERE application_id = $2 AND status = $3;


-- name: ApplicationStatusTo :exec
UPDATE applications
SET status = $1
WHERE application_id = $2;

-- name: InterviewStatusTo :exec
UPDATE interviews
SET status = $1
WHERE application_id = $2;


-- name: ScheduleInterview :exec
INSERT INTO interviews (application_id, company_id, date_time, type, notes, location)
VALUES ($1, (SELECT company_id FROM companies WHERE user_id = $2), $3, $4, $5, $6);


-- name: GetScheduleInterviewData :one
SELECT 
    students.student_name, 
    students.student_email,
    j.title,
    c.company_name
FROM students
JOIN (SELECT job_id, student_id FROM applications WHERE application_id = $1) AS t ON t.student_id = students.student_id
JOIN (SELECT job_id, title, company_id FROM jobs) AS j ON j.job_id = t.job_id
JOIN (SELECT company_id, company_name FROM companies) AS c ON j.company_id = c.company_id;

-- name: GetOfferLetterData :one
SELECT 
    students.student_name, 
    students.student_email,
    j.title,
    c.company_name,
    c.representative_contact,
    c.representative_email
FROM students
JOIN (SELECT job_id, student_id FROM applications WHERE application_id = $1) AS t ON t.student_id = students.student_id
JOIN (SELECT job_id, title, company_id FROM jobs) AS j ON j.job_id = t.job_id
JOIN (SELECT company_id, company_name, representative_contact, representative_email FROM companies) AS c ON j.company_id = c.company_id;


-- name: DeleteInterview :exec
DELETE FROM interviews
WHERE application_id = $1;


-- name: CancelInterviewEmailData :one
SELECT 
    students.student_name, 
    students.student_email,
    j.title,
    c.company_name,
    t.date_time,
    c.representative_name,
    c.representative_email
FROM students
JOIN (
    SELECT 
        job_id, 
        student_id, 
        interviews.date_time
    FROM applications 
    JOIN interviews ON interviews.application_id = applications.application_id 
    WHERE applications.application_id = $1) AS t 
ON t.student_id = students.student_id
JOIN (SELECT job_id, title, company_id FROM jobs) AS j ON j.job_id = t.job_id
JOIN (SELECT company_id, company_name, representative_name, representative_email FROM companies) AS c ON j.company_id = c.company_id;


-- name: GetInterviewsForUserID :many
WITH cte AS (
    SELECT 
        students.student_id
    FROM students
    WHERE students.user_id = $1
)
SELECT  
    companies.company_name,
    jobs.title,
    interviews.interview_id,
    (interviews.date_time)::DATE AS date,
    (interviews.date_time)::TIME::TEXT AS time,
    interviews.type::TEXT,
    interviews.location,
    interviews.notes
FROM applications
JOIN cte ON applications.student_id = cte.student_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN interviews ON interviews.application_id = applications.application_id
JOIN companies ON interviews.company_id = companies.company_id;

-- name: GetTestsForUserID :many
WITH cte AS (
    SELECT  
        applications.job_id,
        jobs.title
    FROM applications
    JOIN jobs ON applications.job_id = jobs.job_id
    WHERE applications.student_id = (SELECT student_id FROM students WHERE students.user_id = $1)
)
SELECT 
    tests.test_id,
    tests.test_name,
    tests.description,
    tests.duration,
    tests.q_count,
    TO_CHAR(tests.end_time, 'YYYY-MM-DD HH24:MI:SS') AS end_time,
    tests.type,    
    companies.company_name,
    cte.title
FROM tests
JOIN cte ON tests.job_id = cte.job_id
JOIN companies ON tests.company_id = companies.company_id
WHERE tests.test_id = $2 OR ($2 = 0);


-- name: NewTest :exec
INSERT INTO tests (test_name, description, duration, q_count, end_time, type, upload_method, job_id, company_id, file_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, (SELECT company_id FROM companies WHERE user_id = $9), $10);


-- name: TakeTest :one
SELECT 
    tests.file_id,
    tests.duration
FROM tests
JOIN applications ON applications.job_id = tests.job_id
WHERE applications.student_id = (SELECT student_id FROM students WHERE user_id = $1)
AND tests.test_id = $2;

-- name: NewTestResult :exec
INSERT INTO testResults (test_id, user_id)
VALUES ($1, $2);