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

-- name: GetUserUUIDFromEmail :one
SELECT 
    users.user_uuid
FROM users 
WHERE users.email = $1;

-- name: GetUserUUIDFromUserID :one
SELECT 
    users.user_uuid
FROM users 
WHERE users.user_id = $1;


-- name: CompanyDashboardData :one
WITH company AS (
    SELECT 
        company_id, 
        company_name,
        representative_name 
    FROM companies WHERE user_id = $1
),
jc AS (
    SELECT COUNT(jobs.job_id) AS jobs_count
    FROM jobs
    WHERE jobs.company_id = (SELECT company_id FROM company)
),
ac AS (
    SELECT COUNT(applications.application_id) AS applications_count
    FROM applications
    JOIN jobs ON applications.job_id = jobs.job_id
    WHERE jobs.company_id = (SELECT company_id FROM company)
),
ic AS (
    SELECT COUNT(interviews.interview_id) AS interviews_count
    FROM interviews
    WHERE interviews.company_id = (SELECT company_id FROM company)
    AND interviews.status = 'Scheduled'
)
SELECT * 
FROM company
CROSS JOIN jc
CROSS JOIN ac
CROSS JOIN ic;


-- name: StudentDashboardData :one
WITH st AS (
    SELECT 
        students.student_id,
        students.student_name
    FROM students
    WHERE students.user_id = $1
),
ac AS (
    SELECT
        COUNT(applications.application_id) AS applications_count
        FROM applications
        WHERE applications.student_id = (SELECT student_id FROM st)
)
SELECT * 
FROM st
CROSS JOIN ac;


-- name: InsertNotifications :exec
INSERT INTO notifications (user_id, title, description, timestamp)
VALUES($1, $2, $3, $4);


-- name: GetNotifications :many
WITH tb AS (
    SELECT
        *
    FROM notifications
    WHERE notifications.user_id = $1
    ORDER BY timestamp DESC
    LIMIT $2
    OFFSET $3
),
upd AS (
    UPDATE notifications
    SET read_status = true
    WHERE notif_id IN (SELECT tb.notif_id FROM tb)
)
SELECT * 
FROM tb;





-- >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
-- Company queries 

-- name: InsertNewJob :exec
INSERT INTO jobs (data_url, company_id, title, location, type, salary, skills, position, extras, description)
VALUES ($1, (SELECT company_id FROM companies WHERE companies.user_id = $2), $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateJob :exec
UPDATE jobs
SET location = $1,
    title = $2,
    description = $3,
    type = $4,
    salary = $5,
    skills = $6,
    position = $7,
    extras = $8
WHERE job_id = $9
AND company_id = (SELECT company_id FROM companies WHERE companies.user_id = $10);





-- name: ExtraInfoCompany :one
INSERT INTO companies (company_name, representative_email, representative_contact, representative_name, data_url, user_id, address, picture_url, website, description, industry)
VALUES ($1, $2, $3, $4, $5, (SELECT user_id FROM users WHERE email = $6), $7, $8, $9, $10, $11)
RETURNING *;


-- name: ExtraInfoStudent :one
INSERT INTO students (student_name, roll_number, student_dob, gender, course, department, year_of_study, resume_url, result_url, cgpa, contact_no, student_email, address, skills, user_id, extras, picture_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, (SELECT user_id FROM users WHERE email = $15), $16, $17)
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
AND (applications.application_id = $3 OR $3 = 0)
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
    COALESCE(t.no_of_applications, 0),
    jobs.description,
    jobs.extras    
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
    WHERE companies.user_id = $1)
ORDER BY jobs.job_id;


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

-- name: ApplicationStatusToAnd :one
WITH upd AS (
    UPDATE applications
    SET status = $1
    WHERE application_id = $2 AND status = $3
    RETURNING student_id
)
SELECT
    students.user_id
FROM students 
JOIN upd ON students.student_id = upd.student_id;

-- name: ApplicationStatusTo :one
WITH upd AS (
    UPDATE applications
    SET status = $1
    WHERE application_id = $2
    RETURNING student_id
)
SELECT
    students.user_id
FROM students 
JOIN upd ON students.student_id = upd.student_id;


-- name: InterviewStatusTo :exec
UPDATE interviews
SET status = $1
WHERE application_id = $2;




-- name: ScheduleInterview :one
INSERT INTO interviews (application_id, company_id, date_time, type, notes, location)
VALUES ($1, (SELECT company_id FROM companies WHERE user_id = $2), $3, $4, $5, $6)
RETURNING TO_CHAR(date_time, 'HH12:MI AM DD-MM-YYYY') AS date_time;


-- name: GetScheduleInterviewData :one
SELECT 
    students.student_name, 
    students.student_email,
    students.user_id,
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
    TO_CHAR(t.date_time, 'HH12:MI AM DD-MM-YYYY') AS date_time,
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


-- name: UpcomingInterviewsStudent :many
SELECT 
    companies.company_name,
    jobs.title,
    interviews.interview_id,
    TO_CHAR(interviews.date_time, 'HH12:MI AM DD-MM-YYYY') AS date_time,
    interviews.type::TEXT,
    interviews.location,
    interviews.notes
FROM applications
JOIN interviews ON applications.application_id = interviews.application_id 
                AND interviews.status != 'Completed'
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE applications.student_id = (SELECT student_id FROM students WHERE students.user_id = $1)
AND interviews.date_time > NOW();

-- name: UpcomingTestsStudent :many
SELECT 
    tests.test_id,
    tests.test_name,
    tests.description,
    tests.duration,
    tests.q_count,
    TO_CHAR(tests.end_time, 'HH12:MI AM DD-MM-YYYY') AS end_time,
    tests.type,    
    companies.company_name,
    jobs.title
FROM applications 
JOIN tests ON applications.job_id = tests.job_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE applications.student_id = (SELECT student_id FROM students WHERE students.user_id = $1)
AND NOT EXISTS (SELECT 1 FROM testresults WHERE testresults.test_id = tests.test_id AND testresults.user_id = $1)
AND tests.end_time > NOW();


-- name: ScheduledInterviewsCompany :many
SELECT 
    interviews.interview_id,
    TO_CHAR(interviews.date_time, 'HH12:MI AM DD-MM-YYYY') AS date_time,
    interviews.type::TEXT,
    interviews.status::TEXT,
    interviews.notes,
    interviews.location,
    interviews.extras,
    applications.application_id,
    applications.job_id,
    students.student_id,
    students.student_name,
    students.roll_number
FROM interviews
JOIN applications ON interviews.application_id = applications.application_id
JOIN students ON applications.student_id = students.student_id
WHERE interviews.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1)
AND interviews.status = 'Scheduled' AND interviews.date_time > NOW()
ORDER BY interviews.date_time;

-- name: ScheduledTestsCompany :many
SELECT 
    tests.test_id,
    tests.test_name,
    tests.description,
    tests.duration,
    tests.q_count,
    TO_CHAR(tests.end_time, 'HH12:MI AM DD-MM-YYYY') AS end_time,
    tests.type,
    tests.threshold,
    jobs.job_id,
    jobs.title
FROM tests
JOIN jobs ON tests.job_id = jobs.job_id
WHERE tests.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1)
AND tests.end_time > NOW()
ORDER BY tests.end_time;





-- name: CompletedTestsStudent :many
SELECT 
    tests.test_id,
    tests.test_name,
    tests.published,

    testresults.result_id,
    TO_CHAR(testresults.start_time, 'HH12:MI AM DD-MM-YYYY') AS start_time,
    TO_CHAR(testresults.end_time, 'HH12:MI AM DD-MM-YYYY') AS end_time,

    companies.company_name,
    jobs.title
FROM testresults
JOIN tests ON testresults.test_id = tests.test_id
JOIN companies ON tests.company_id = companies.company_id
JOIN jobs ON tests.job_id = jobs.job_id
WHERE testresults.user_id = $1
AND testresults.end_time IS NOT NULL
ORDER BY testresults.result_id;

-- name: CompletedInterviewsStudent :many
SELECT 
    jobs.title AS job_title,
    companies.company_name,
    interviews.interview_id,
    interviews.application_id,
    TO_CHAR(interviews.date_time, 'HH12:MI AM DD-MM-YYYY') AS interview_date_time,
    interviews.extras,


    (CASE WHEN feedback_id IS NULL THEN false ELSE true END) AS feedback_given,
    feedbacks.feedback_id,
    feedbacks.message AS feedback_message
FROM interviews
JOIN applications ON interviews.application_id = applications.application_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
LEFT JOIN feedbacks ON (feedbacks.interview_id = interviews.interview_id AND feedbacks.user_id = $1) 
WHERE applications.student_id = (SELECT student_id FROM students WHERE students.user_id = $1)
AND interviews.status = 'Completed';



-- name: CompletedTestsCompany :many
SELECT 
    tests.test_id,
    tests.test_name,
    TO_CHAR(tests.end_time, 'HH12:MI AM DD-MM-YYYY') AS end_time,
    tests.threshold,
    tests.published,
    TO_CHAR(tests.created_at, 'HH12:MI AM DD-MM-YYYY') AS created_at
FROM tests
WHERE tests.end_time < NOW()
AND tests.company_id = (SELECT company_id FROM companies WHERE companies.user_id = $1);

-- name: CompletedInterviewsCompany :many
SELECT 
    students.student_name,
    students.student_id,
    interviews.interview_id,
    interviews.application_id,
    TO_CHAR(interviews.date_time, 'HH12:MI AM DD-MM-YYYY') AS date_time,
    interviews.extras,

    (CASE WHEN feedback_id IS NULL THEN false ELSE true END) AS feedback_given,
    feedbacks.feedback_id,
    feedbacks.message AS feedback_message
FROM interviews
JOIN applications ON interviews.application_id = applications.application_id
JOIN students ON students.student_id = applications.student_id
LEFT JOIN feedbacks ON (feedbacks.interview_id = interviews.interview_id AND feedbacks.user_id = $1)
WHERE interviews.company_id = (SELECT company_id FROM companies WHERE companies.user_id = $1)
AND interviews.status = 'Completed';






-- name: TestMetadata :one
SELECT 
    tests.test_id,
    tests.test_name,
    tests.description,
    tests.duration,
    tests.q_count,
    TO_CHAR(tests.end_time, 'HH12:MI AM DD-MM-YYYY') AS end_time,
    tests.type,    
    companies.company_name,
    jobs.title
FROM applications
JOIN tests ON applications.job_id = tests.job_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE applications.student_id = (SELECT student_id FROM students WHERE students.user_id = $1)
AND tests.test_id = $2;


-- name: NewTest :exec
INSERT INTO tests (test_name, description, duration, q_count, end_time, type, upload_method, job_id, company_id, file_id, threshold)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, (SELECT company_id FROM companies WHERE user_id = $9), $10, $11);


-- name: TakeTest :one
SELECT 
    tests.file_id,
    tests.duration,
    tests.end_time
FROM tests
JOIN applications ON applications.job_id = tests.job_id
WHERE applications.student_id = (SELECT student_id FROM students WHERE user_id = $1)
AND tests.test_id = $2;

-- name: NewTestResult :exec
INSERT INTO testresults (test_id, user_id, start_time)
VALUES ($1, $2, $3);

-- name: IsTestGiven :one
SELECT 
    testresults.result_id,
    testresults.test_id,
    testresults.start_time,
    testresults.end_time
FROM testresults
WHERE testresults.test_id = $1 
AND testresults.user_id = $2;

-- name: UpdateResponse :exec
INSERT INTO testresponses (result_id, question_id, response, time_taken)
VALUES ($1, $2, $3, $4)
ON CONFLICT (result_id, question_id)
DO UPDATE SET response = $3, time_taken = $4;


-- name: SubmitTest :one
UPDATE testresults 
SET end_time = $1
WHERE user_id = $2 
AND test_id = $3
RETURNING result_id;


-- name: StudentProfileData :one
SELECT 
    students.student_name,
    students.roll_number,
    students.student_dob,
    students.gender,
    students.course,
    students.department,
    students.year_of_study,
    students.cgpa,
    students.contact_no,
    students.student_email,
    students.address,
    students.skills,
    students.extras
FROM students
WHERE students.user_id = $1;


-- name: CompanyProfileData :one
SELECT 
    companies.company_name,
    companies.representative_email,
    companies.representative_contact,
    companies.representative_name,
    companies.address,
    companies.website,
    companies.description,
    companies.industry
FROM companies
WHERE companies.user_id = $1;

-- name: TestHistory :many
SELECT 
    tests.test_id,
    tests.test_name,
    TO_CHAR(testresults.start_time, 'HH12:MI AM DD-MM-YYYY') AS start_time
FROM testresults
JOIN tests ON testresults.test_id = tests.test_id
WHERE testresults.user_id = $1
ORDER BY testresults.start_time DESC;

-- name: ApplicationHistory :many
SELECT 
    applications.application_id,
    TO_CHAR(applications.created_at, 'HH12:MI AM DD-MM-YYYY') AS created_at,
    jobs.title,
    companies.company_name
FROM applications
JOIN jobs ON jobs.job_id = applications.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE applications.student_id = (SELECT students.student_id FROM students WHERE students.user_id = $1)
ORDER BY applications.created_at DESC;

-- name: InterviewHistory :many
SELECT 
    interviews.interview_id,
    TO_CHAR(interviews.created_at, 'HH12:MI AM DD-MM-YYYY') AS created_at,
    jobs.title
FROM interviews
JOIN applications ON applications.application_id = interviews.application_id
JOIN jobs ON jobs.job_id = applications.job_id
WHERE applications.student_id = (SELECT students.student_id FROM students WHERE students.user_id = $1)
ORDER BY interviews.created_at DESC;

-- name: ApplicationsStatusCounts :one
SELECT
    CAST(COALESCE(COUNT(applications.status), 0) AS BIGINT) AS applied_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'UnderReview' THEN 1 END), 0) AS BIGINT) AS under_review_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'ShortListed' THEN 1 END), 0) AS BIGINT) AS shortlisted_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'Rejected' THEN 1 END), 0) AS BIGINT) AS rejected_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'Offered' THEN 1 END), 0) AS BIGINT) AS offered_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'Hired' THEN 1 END), 0) AS BIGINT) AS hired_count
FROM applications
WHERE applications.student_id = (SELECT students.student_id FROM students WHERE students.user_id = $1);

-- name: GetAllJobsID :many
SELECT
    jobs.job_id
FROM jobs
WHERE jobs.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1);

-- name: ApplicantsCount :many
WITH ji AS (
    SELECT
        jobs.job_id
    FROM jobs
    WHERE jobs.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1)
)

SELECT
    applications.job_id,
    CAST(COALESCE(COUNT(applications.status), 0) AS BIGINT) AS total_apps,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'UnderReview' THEN 1 END), 0) AS BIGINT) AS reviewed_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'ShortListed' THEN 1 END), 0) AS BIGINT) AS shortlisted_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'Rejected' THEN 1 END), 0) AS BIGINT) AS rejected_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'Offered' THEN 1 END), 0) AS BIGINT) AS offered_count,
    CAST(COALESCE(SUM(CASE WHEN applications.status = 'Hired' THEN 1 END), 0) AS BIGINT) AS hired_count
FROM applications
JOIN ji ON ji.job_id = applications.job_id
GROUP BY applications.job_id;

-- name: UsersTableData :one
SELECT 
    TO_CHAR(users.created_at, 'HH12:MI AM DD-MM-YYYY') AS created_at,
    users.confirmed,
    users.is_verified
FROM users
WHERE users.user_id = $1;


-- name: GetAllFilePathsStudent :one
SELECT 
    students.resume_url,
    students.result_url,
    students.picture_url
FROM students
WHERE user_id = $1;

-- name: GetAllFilePathsCompany :one
SELECT 
    companies.picture_url,
    companies.picture_url
FROM companies
WHERE user_id = $1;



-- name: UpdateStudentDetails :exec
UPDATE students
SET course = $1,
    department = $2,
    year_of_study = $3,
    cgpa = $4,
    contact_no = $5,
    address = $6,
    skills = $7
WHERE user_id = $8;

-- name: UpdateCompanyDetails :exec
UPDATE companies
SET company_name = $1,
    representative_email = $2,
    representative_contact = $3,
    representative_name = $4,
    address = $5,
    website = $6,
    description = $7,
    industry = $8
WHERE user_id = $9;


-- name: UpdateStudentResume :exec
UPDATE students
SET
    resume_url = $1
WHERE user_id = $2;

-- name: UpdateStudentResult :exec
UPDATE students
SET
    result_url = $1
WHERE user_id = $2;

-- name: UpdateStudentProfilePic :exec
UPDATE students
SET
    picture_url = $1
WHERE user_id = $2;

-- name: UpdateCompanyProfilePic :exec
UPDATE companies
SET
    picture_url = $1
WHERE user_id = $2;

-- name: IsTestPublished :one
SELECT 
    tests.published
FROM tests
WHERE tests.test_id = $1;

-- name: UpdateTest :exec
UPDATE tests 
SET 
    threshold = COALESCE(sqlc.narg('threshold'), threshold),
    published = COALESCE(sqlc.narg('published'), published)
WHERE tests.test_id = $1
AND tests.company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $2);


-- name: UpdateInterview :one
UPDATE interviews
SET 
    date_time = $3,
    type = $4,
    notes = $5,
    location = $6
WHERE company_id = (SELECT companies.company_id FROM companies WHERE companies.user_id = $1)
AND interview_id = $2
RETURNING application_id, TO_CHAR(date_time, 'HH12:MI AM DD-MM-YYYY') AS date_time;



-- name: TestResultPoller :one
SELECT  
    tests.test_id
FROM tests
WHERE tests.end_time < NOW()
AND tests.result_url IS NULL
LIMIT 1;

-- name: TestAuthorization :one
SELECT 
    tests.test_id
FROM tests
WHERE tests.test_id = $1
AND tests.company_id = (SELECT company_id FROM companies WHERE companies.user_id = $2)
AND tests.end_time < NOW();

-- name: TestData :one
SELECT 
    tests.file_id,
    tests.test_id,
    tests.test_name,
    tests.q_count,
    TO_CHAR(tests.end_time, 'HH12:MI AM DD-MM-YYYY') AS end_time,
    tests.threshold,
    jobs.title,
    companies.company_name,
    companies.representative_email
FROM tests
JOIN companies ON tests.company_id = companies.company_id
JOIN jobs ON tests.job_id = jobs.job_id
WHERE tests.test_id = $1;

-- name: ClearAnswersTable :exec
DELETE FROM temp_correct_answers;

-- name: InsertAnswers :exec
INSERT INTO temp_correct_answers (question_id, correct_answer, points)
VALUES ($1, $2, $3);

-- name: UpdateTestResultURLUnprotected :exec
UPDATE tests
SET 
    result_url = $2
WHERE test_id = $1;

-- name: EvaluateTestResult :one
WITH tr AS (
    UPDATE testresponses
    SET points = temp_correct_answers.points
    FROM temp_correct_answers
    WHERE testresponses.question_id = temp_correct_answers.question_id
    AND testresponses.response = temp_correct_answers.correct_answer
    RETURNING testresponses.points, testresponses.result_id
),
rs AS (
    SELECT 
        result_id, 
        SUM(points) AS score 
    FROM tr
    GROUP BY result_id
),
up AS (
    UPDATE testresults
    SET score = rs.score
    FROM rs
    WHERE testresults.result_id = rs.result_id
)
SELECT 
    SUM(temp_correct_answers.points) AS totalpoints
FROM temp_correct_answers;

-- name: CumulativeResultData :many
WITH tr AS (
    SELECT 
        result_id,
        SUM(time_taken) AS total_time_taken,
        COUNT(result_id) AS questions_attempted
    FROM testresponses
    GROUP BY result_id
),
main AS (
    SELECT 
        testresults.result_id,
        testresults.user_id,
        TO_CHAR(testresults.start_time, 'HH12:MI:SS AM DD-MM-YYYY') AS start_time,
        TO_CHAR(testresults.end_time, 'HH12:MI:SS AM DD-MM-YYYY') AS end_time,
        testresults.score,
        tr.total_time_taken,
        tr.questions_attempted
    FROM testresults 
    JOIN tr ON testresults.result_id = tr.result_id
    WHERE testresults.test_id = $1
)
SELECT * 
FROM main;

-- name: TestPassFailCount :one
SELECT
    COUNT(CASE WHEN testresults.score >= $2 THEN 1 ELSE NULL END) AS pass_count,
    COUNT(CASE WHEN testresults.score < $2 THEN 1 ELSE NULL END) AS fail_count
FROM testresults
WHERE testresults.test_id = $1;

-- name: StudentTestResult :many
WITH tr AS (
    SELECT 
        result_id,
        SUM(time_taken) AS total_time_taken,
        COUNT(result_id) AS questions_attempted,
        COUNT(CASE WHEN points > 0 THEN 1 ELSE NULL END) AS correct_response
    FROM testresponses
    GROUP BY result_id
)
SELECT 
    testresults.result_id,
    TO_CHAR(testresults.start_time, 'HH12:MI:SS AM DD-MM-YYYY') AS start_time,
    TO_CHAR(testresults.end_time, 'HH12:MI:SS AM DD-MM-YYYY') AS end_time,
    testresults.score,

    students.student_id,
    students.student_name,
    students.roll_number,
    students.student_email,

    tr.total_time_taken,
    tr.questions_attempted,
    tr.correct_response,

    users.user_uuid
FROM testresults
JOIN students ON testresults.user_id = students.user_id
JOIN users ON testresults.user_id = users.user_id
JOIN tr ON testresults.result_id = tr.result_id
WHERE testresults.test_id = $1;






















-- name: InsertFeedbackByCompanyToStudent :exec
INSERT INTO feedbacks (application_id, interview_id, user_id, message)
VALUES ($1, $2, $3, $4); 

-- name: InsertFeedbackByStudentToCompany :exec
INSERT INTO feedbacks (application_id, interview_id, user_id, message)
VALUES ($1, $2, $3, $4); 





-- name: FeedbacksByCompanyUserToStudents :many
SELECT 
    feedbacks.feedback_id,
    TO_CHAR(feedbacks.created_at, 'HH12:MI:SS AM DD-MM-YYYY') AS feedback_time,
    feedbacks.message,
    feedbacks.application_id,
    feedbacks.interview_id,

    (CASE WHEN feedbacks.application_id IS NULL THEN false ELSE true END) AS application_feedback,
    (CASE WHEN feedbacks.interview_id IS NULL THEN false ELSE true END) AS interview_feedback,

    students.student_id,
    students.student_name
FROM feedbacks
LEFT JOIN interviews ON interviews.interview_id = feedbacks.interview_id
LEFT JOIN applications ON (applications.application_id = feedbacks.application_id OR 
                            applications.application_id = interviews.application_id)
JOIN students ON students.student_id = applications.student_id
WHERE feedbacks.user_id = $1
ORDER BY feedbacks.created_at DESC;

  
-- name: FeedbacksByStudentsToCompanyUser :many
WITH notuser AS (
    SELECT 
        interviews.interview_id 
    FROM interviews 
    WHERE interviews.company_id = 
        (SELECT companies.company_id FROM companies WHERE companies.user_id = $1)
)
SELECT 
    feedbacks.feedback_id,
    TO_CHAR(feedbacks.created_at, 'HH12:MI:SS AM DD-MM-YYYY') AS feedback_time,
    feedbacks.message
FROM feedbacks
LEFT JOIN notuser ON notuser.interview_id = feedbacks.interview_id
WHERE feedbacks.user_id != $1 AND notuser.interview_id IS NOT NULL
ORDER BY feedbacks.created_at DESC;
  













-- name: FeedbacksByCompaniesToStudent :many
SELECT 
    feedbacks.feedback_id,
    TO_CHAR(feedbacks.created_at, 'HH12:MI:SS AM DD-MM-YYYY') AS feedback_time,
    feedbacks.application_id,
    feedbacks.interview_id,
    feedbacks.message,
    (CASE WHEN feedbacks.application_id IS NULL THEN false ELSE true END) AS application_feedback,
    (CASE WHEN feedbacks.interview_id IS NULL THEN false ELSE true END) AS interview_feedback,

    jobs.title AS job_title,
    companies.company_name
FROM feedbacks
LEFT JOIN interviews ON feedbacks.interview_id = interviews.interview_id
LEFT JOIN applications ON (interviews.application_id = applications.application_id OR 
                            feedbacks.application_id = applications.application_id)
JOIN students ON applications.student_id = students.student_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE feedbacks.user_id != $1 AND students.user_id = $1
ORDER BY feedbacks.created_at DESC;

-- name: FeedbacksByStudentToCompanies :many
SELECT 
    feedbacks.feedback_id,
    TO_CHAR(feedbacks.created_at, 'HH12:MI:SS AM DD-MM-YYYY') AS feedback_time,
    feedbacks.application_id,
    feedbacks.interview_id,
    feedbacks.message,
    (CASE WHEN feedbacks.application_id IS NULL THEN false ELSE true END) AS application_feedback,
    (CASE WHEN feedbacks.interview_id IS NULL THEN false ELSE true END) AS interview_feedback,

    jobs.title AS job_title,
    companies.company_name
FROM feedbacks
LEFT JOIN interviews ON feedbacks.interview_id = interviews.interview_id
LEFT JOIN applications ON (interviews.application_id = applications.application_id OR 
                            feedbacks.application_id = applications.application_id)
JOIN students ON applications.student_id = students.student_id
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
WHERE feedbacks.user_id = $1
ORDER BY feedbacks.created_at DESC;
























-- name: DiscussionsData :many
SELECT 
    discussions.post_id,

    discussions.content,
    TO_CHAR(discussions.created_at, 'HH12:MI:SS AM DD-MM-YYYY') AS created_at,

    companies.company_name,
    students.student_name,

    (CASE WHEN discussions.user_id = $3 THEN true ELSE false END) AS owned

FROM discussions 
LEFT JOIN students ON discussions.user_id = students.user_id
LEFT JOIN companies ON discussions.user_id = companies.user_id
WHERE thread_id IS NULL
ORDER BY discussions.created_at DESC
OFFSET $1 LIMIT $2;

-- name: InsertDiscussion :exec
INSERT INTO discussions (user_id, role, content)
VALUES ($1, (SELECT role FROM users WHERE users.user_id = $1), $2);

-- name: UpdateDiscussion :exec
UPDATE discussions
SET
    content = $3
WHERE post_id = $2
AND user_id = $1;

-- name: GetReplies :many
SELECT 
    TO_CHAR(discussions.created_at, 'HH12:MI:SS AM DD-MM-YYYY') AS created_at,
    discussions.content,

    companies.company_name,
    students.student_name

FROM discussions 
LEFT JOIN students ON discussions.user_id = students.user_id
LEFT JOIN companies ON discussions.user_id = companies.user_id
WHERE discussions.thread_id = $1;
















-- name: StudentProfileForCompany :one
SELECT
    students.student_name,
    students.roll_number,
    students.student_dob,
    students.gender,
    students.course,
    students.department,
    students.year_of_study,
    students.cgpa,
    students.contact_no,
    students.student_email,
    students.address,
    students.skills,
    students.extras
FROM applications
JOIN jobs ON applications.job_id = jobs.job_id
JOIN companies ON jobs.company_id = companies.company_id
JOIN students ON applications.student_id = students.student_id
WHERE applications.student_id = $1 AND companies.user_id = $2;

-- -- name: GetStudentProfile :one
-- SELECT
--     students.student_name,
--     students.roll_number,
--     students.student_dob,
--     students.gender,
--     students.course,
--     students.department,
--     students.year_of_study,
--     students.cgpa,
--     students.contact_no,
--     students.student_email,
--     students.address,
--     students.skills,
--     students.extras
-- FROM students
-- WHERE students.student_id = $1;

-- >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
-- Admin Functions --------------------------------



-- name: ListToVerifyStudent :many
SELECT 
    users.user_id,
    users.email,
    TO_CHAR(users.created_at, 'HH12:MI AM DD-MM-YYYY') AS created_at,
    users.confirmed
FROM users
WHERE users.confirmed = true
AND users.is_verified = false
AND users.role = 1;

-- name: VerifyStudent :exec
UPDATE users
SET is_verified = true
WHERE user_id = $1;

-- name: StudentsOverview :many
SELECT 
    students.student_id,
    students.student_name,
    students.roll_number,
    students.student_dob,
    students.gender,
    students.course,
    students.department,
    students.year_of_study,
    students.cgpa,
    students.contact_no,
    students.address,
    students.skills, 
    students.extras
FROM students;

-- name: StudentInfo :one
SELECT
    students.student_id,
    students.student_name,
    students.roll_number,
    students.student_dob,
    students.gender,
    students.course,
    students.department,
    students.year_of_study,
    students.cgpa,
    students.contact_no,
    students.address,
    students.skills, 
    students.picture_url AS profilePic,
    students.extras
FROM students
WHERE students.user_id = $1;









