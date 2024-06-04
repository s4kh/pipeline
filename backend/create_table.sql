CREATE TABLE votes (
  candidate_id VARCHAR (50) PRIMARY KEY,
  candidate_name varchar(50),
  party_id VARCHAR(50) NOT NULL, 
  party_name VARCHAR(50),
  total_vote INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);