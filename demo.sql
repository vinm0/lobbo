CREATE TABLE leaders (
	leader_id	INTEGER PRIMARY KEY AUTOINCREMENT,
	usrname	    TEXT NOT NULL,
	pwd	        TEXT NOT NULL,
	fname		TEXT NOT NULL,
	lname		TEXT NOT NULL
);

CREATE TABLE lobby (
	lobby_id	INTEGER PRIMARY KEY AUTOINCREMENT,
	owner_id	INTEGER NOT NULL,
	title		TEXT NOT NULL,
	lobby_desc	TEXT,
	meet_time	TEXT,
	meet_loc	TEXT,
	capacity	INTEGER,
	visibility	INTEGER DEFAULT 0,
	invite_only	INTEGER DEFAULT 0,
		CONSTRAINT lobby_owner_fk FOREIGN KEY (owner_id) 
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE
);

CREATE TABLE lobby_members (
	lobby_id	INTEGER,
	member_id	INTEGER,
		PRIMARY KEY (lobby_id, member_id),
		CONSTRAINT lobby_members_lobby_fk FOREIGN KEY (lobby_id) 
			REFERENCES lobby(lobby_id) 
			ON DELETE CASCADE,
		CONSTRAINT lobby_members_member_fk FOREIGN KEY (member_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE
);

CREATE TABLE colleagues (
	owner_id		INTEGER,
	colleague_id	INTEGER,
		PRIMARY KEY (owner_id, colleague_id),
		CONSTRAINT colleagues_owner_fk FOREIGN KEY (owner_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE,
		CONSTRAINT colleagues_colleague_fk FOREIGN KEY (colleague_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE
);

CREATE TABLE profiles (
	owner_id	INTEGER PRIMARY KEY,
	email		TEXT NOT NULL,
	bio			TEXT,
		CONSTRAINT profiles_owner_fk FOREIGN KEY (owner_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE 
);

CREATE TABLE preferences (
	owner_id 	INTEGER PRIMARY KEY,
	visibility	INTEGER DEFAULT 0,
		CONSTRAINT preferences_owner_fk FOREIGN KEY (owner_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE 
);


CREATE TABLE requests (
	request_id		INTEGER PRIMARY KEY,
	sender_id		INTEGER NOT NULL,
	receiver_id		INTEGER NOT NULL,
	req_type		TEXT NOT NULL,
	send_date		TEXT NOT NULL,
	response_date	TEXT,
	response		INTEGER,
		CONSTRAINT requests_sender_fk FOREIGN KEY (sender_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE,
		CONSTRAINT requests_receiver_fk FOREIGN KEY (receiver_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE 
);