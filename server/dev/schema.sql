
CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       role VARCHAR(50) NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       verified BOOLEAN NOT NULL DEFAULT FALSE,
                       verification_token varchar(255),
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       deleted_at TIMESTAMP NOT NULL
);

CREATE TABLE user_sessions (
                               user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                               token TEXT NOT NULL,
                               refresh_token TEXT NOT NULL,
                               PRIMARY KEY (user_id)
);

CREATE TABLE chat_rooms (
                            id VARCHAR(255) PRIMARY KEY,
                            name VARCHAR(255) NOT NULL,
                            members JSONB NOT NULL
);

CREATE TABLE messages (
                          id UUID PRIMARY KEY,
                          chat_room_id varchar(255) REFERENCES chat_rooms(id) ON DELETE CASCADE,
                          author_id UUID REFERENCES users(id) ON DELETE CASCADE,
                          text TEXT NOT NULL,
                          seen_by JSONB NOT NULL,
                          received_by JSONB NOT NULL,
                          sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          deleted_at TIMESTAMP
);
verification_token varchar(255)