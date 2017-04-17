CREATE TABLE admusers (
  id SERIAL PRIMARY KEY,
  username VARCHAR(100) UNIQUE,
  email_address VARCHAR(100),
  password VARCHAR(100),
  salt VARCHAR(100)
);

CREATE TABLE emails (
  id SERIAL PRIMARY KEY,
  from_addr VARCHAR(255),
  from_display VARCHAR(255),
  subject VARCHAR(500),
  plain_text TEXT,
  formatted_text TEXT,
  is_read BOOLEAN DEFAULT(false),
  is_spam BOOLEAN DEFAULT(false),
  is_virus BOOLEAN DEFAULT(false),
  received TIMESTAMP WITH TIME ZONE
);

CREATE TABLE recipients (
  email_id INT REFERENCES emails(id) ON DELETE CASCADE,
  email_address VARCHAR(255),
  recipient_type VARCHAR(4)
);

CREATE TABLE attachments (
  id SERIAL PRIMARY KEY,
  content_type VARCHAR(255),
  file_name VARCHAR(255),
  file_path VARCHAR(1000),
  email_id INT REFERENCES emails(id) ON DELETE CASCADE
);

/*
DROP TABLE admusers;
DROP TABLE attachments;
DROP TABLE recipients;
DROP TABLE emails;
*/
