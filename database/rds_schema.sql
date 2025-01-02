CREATE TABLE res_colleges (
    name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO res_colleges 
    (name) 
VALUES 
    ('Timothy Dwight'),
    ('Silliman'),
    ('Berkeley'),
    ('Branford'),
    ('Saybrook'),
    ('Davenport'),
    ('Ezra Stiles'),
    ('Morse'),
    ('Pauli Murray'),
    ('Benjamin Franklin'),
    ('Grace Hopper'),
    ('Jonathan Edwards'),
    ('Pierson'),
    ('Trumbull');

/* 
   gender will be stored as int, as the binary representation of the int 
   mapping to booleans representing each gender. 
   ex: (int) 4 >> (binary) 00100 >> (boolean) [false, false, true, false, false] & 
   [cis-female, trans-female, cis-male, trans-male, non-binary] = cis-male
*/

CREATE TABLE test (
    name VARCHAR(20) PRIMARY KEY
);

INSERT INTO test
    (name)
VALUES
    ('firstname');

CREATE TABLE users (
    email VARCHAR(50) PRIMARY KEY,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(50) NOT NULL,
    residential_college VARCHAR(30) REFERENCES res_colleges(name),
    graduating_year INT,
    gender INT, -- see above
    partner_genders INT, -- see above
    instagram VARCHAR(30),
    snapchat VARCHAR(30),
    phone_number VARCHAR(15),
    picture_s3_url VARCHAR(2000),
    interest_1 VARCHAR(20),
    interest_2 VARCHAR(20),
    interest_3 VARCHAR(20),
    interest_4 VARCHAR(20),
    interest_5 VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE answers (
    email VARCHAR(50) PRIMARY KEY REFERENCES users(email),
    question1 INT,
    question2 INT,
    question3 INT,
    question4 INT,
    question5 INT,
    question6 INT,
    question7 INT,
    question8 INT,
    question9 INT,
    question10 INT,
    question11 INT,
    question12 INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE elos (
    email VARCHAR(50) PRIMARY KEY REFERENCES users(email),
    elo INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE matches (
    user1_email VARCHAR(50) REFERENCES users(email),
    user2_email VARCHAR(50) REFERENCES users(email),
    user1_interested BOOLEAN NOT NULL DEFAULT FALSE,
    user2_interested BOOLEAN NOT NULL DEFAULT FALSE,
    server_generated BOOLEAN NOT NULL DEFAULT FALSE,
    week TIMESTAMP NOT NULL, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user1_email, user2_email)
);

CREATE OR REPLACE FUNCTION create_answers_row()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO answers (email) VALUES (NEW.email);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_answers_row
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_answers_row();

CREATE OR REPLACE FUNCTION create_elos_row()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO elos (email) VALUES (NEW.email);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_elos_row
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_elos_row();

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON answers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON elos
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON matches
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_matches_week ON matches (week);




