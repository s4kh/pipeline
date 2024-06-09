CREATE TABLE candidate_votes (
  candidate_id VARCHAR (50) PRIMARY KEY,
  candidate_name varchar(50),
  party_id VARCHAR(50) NOT NULL, 
  total_vote INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE TABLE party_votes (
  party_id VARCHAR (50) PRIMARY KEY,
  total_vote INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);