CREATE TABLE leaders (
	leader_id	INTEGER PRIMARY KEY AUTOINCREMENT,
	usrname	    TEXT NOT NULL,
	pwd	        TEXT NOT NULL,
	fname		TEXT NOT NULL,
	lname		TEXT NOT NULL,
		UNIQUE(usrname)
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
		CHECK(owner_id != colleague_id)
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
	visibility	INTEGER DEFAULT 'public',
		CHECK (visibility IN (
				'public', 
				'user', 
				'friend of friend', 
				'friend', 
				'private'))
		CONSTRAINT preferences_owner_fk FOREIGN KEY (owner_id)
			REFERENCES leaders(leader_id)
			ON DELETE CASCADE 
);


CREATE TABLE requests (
	request_id		INTEGER PRIMARY KEY AUTOINCREMENT,
	sender_id		INTEGER NOT NULL,
	reference_id	INTEGER NOT NULL
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

INSERT INTO leaders (usrname, pwd, fname, lname)
VALUES
	("jaynorth", "sledderconcerned", "Jason", "Northwood"),
	("painterlaveer", "differangolan", "Pam", "Leer"),
	("homesicksilk", "plinkponie", "Homer", "Slick"),
	("toothquality", "possepeg", "Tabitha", "Lity"),
	("poundbarely", "beautyelytra", "Pablo", "Barley"),
	("supremeassertive", "velcroexhibition", "Savanna", "Serti"),
	("chapterwitness", "wereenable", "Chap", "Wit"),
	("toldearn", "sandpiperkey", "Todd", "Larn"),
	("tagfuture", "whipstaffpastie", "Anselm", "Future"),
	("edibleskit", "profitnancy", "Debbie", "Kitt")
;

INSERT INTO lobby (	
	owner_id,
	title,
	lobby_desc,
	meet_time,
	meet_loc,
	capacity,
	visibility,
	invite_only)
VALUES
	(1, "beauty", NULL, "2021-01-20 00:00:00", 
		"Zoom", 10, 0, 0),
	(2, "profit", "Look at company numbers", "2021-07-24 00:00:00", 
		"123 N Main Ave", 5, 4, 1),
	(3, "networking", NULL, "2021-07-25 00:00:00", 
		"345 Mulberry", 50, 0, 0),
	(4, "scrubs", "Nurses meeting", "2021-08-01 00:00:00", 
		"Staff Lounge", 15, 2, 0),
	(1, "chapter 9", "Book club meeting", "2021-10-31 00:00:00", 
		"Marcy's House", 10, 1, 1),
	(3, "homesick", NULL, "2021-01-10 00:00:00", 
		"Mama's", 3, 4, 1),
	(2, "supreme", "AHS fan club", "2021-03-19 00:00:00", 
		"New Orleans, LA", NULL, 0, 0),
	(2, "skit", "Improv night", "2021-09-18 00:00:00", 
		"Community Theater", 30, 0, 0),
	(7, "yelp", "Get reviews up", "2021-08-21 00:00:00", 
		"Cafe", 10, 1, 0),
	(6, "vegan", "Vegans anonymous", "2021-01-30 00:00:00", 
		"Greens Restaurant", 15, 2, 1)
;

INSERT INTO lobby_members (lobby_id, member_id) 
VALUES
	(7, 1), (7, 4),
	(7, 3), (7, 6),
	(7, 5), (7, 8),
	(7, 7), (3, 1),
	(7, 9), (3, 2),
	(3, 7), (3, 4),
	(1, 9), (3, 5),
	(1, 8), (3, 6),
	(1, 7), (1, 6),
	(4, 5), (4, 3),
	(4, 9), (4, 7),
	(5, 2), (4, 1),
	(5, 6), (5, 4),
	(6, 1), (6, 2),
	(8, 6), (8, 9),
	(8, 1), (8, 8),
	(8, 3), (8, 7),
	(9, 4), (9, 8),
	(9, 1), (9, 6),
	(9, 5), (9, 3),
	(9, 2), (9, 9)
;

INSERT INTO colleagues (owner_id, colleague_id)
VALUES
	(1, 9), (9, 1),
	(1, 7), (9, 3),
	(1, 4), (9, 5),
	(2, 3), (8, 1),
	(2, 5), (8, 5),
	(3, 7), (7, 9),
	(3, 9), (7, 6),
	(4, 6), (6, 4),
	(4, 1), (6, 2),
	(10, 1), (10, 2)
;


INSERT INTO profiles (owner_id, email, bio)
VALUES
	(1, "ukxxv@gmail.com", ""),
	(2, "nqkah@gmail.com", ""),
	(3, "murrr@gmail.com", ""),
	(4, "abbsk@gmail.com", ""),
	(5, "wrncj@gmail.com", ""),
	(6, "zeqrq@gmail.com", ""),
	(7, "waauq@gmail.com", ""),
	(8, "jrncf@gmail.com", ""),
	(9, "twhav@gmail.com", ""),
	(10, "qzmfa@gmail.com", "")
;


INSERT INTO preferences (owner_id, visibility)
VALUES
	(1, 'public'),
	(2, 'user'),
	(3, 'public'),
	(4, 'user'),
	(5, 'friend'),
	(6, 'public'),
	(7, 'friend of friend'),
	(8, 'public'),
	(9, 'public'),
	(10, 'private')
;


INSERT INTO requests (
	sender_id, 
	receiver_id, 
	reference_id,
	req_type, 
	send_date, 
	response, 
	response_date)
VALUES
	(1, 8, 5, "l invite", "2021-07-10 00:00:00", NULL, NULL),
	(2, 8, 2, "l join", "2021-07-10 00:00:00", NULL, NULL),
	(2, 7, 2, "l invite", "2021-07-10 00:00:00", NULL, NULL),
	(2, 6, 2, "l invite", "2021-07-10 00:00:00", NULL, NULL),
	(4, 1, 4, "l join", "2021-07-10 00:00:00", NULL, NULL),
	(6, 4, 10, "l invite", "2021-07-10 00:00:00", NULL, NULL),
	(6, 3, 10, "l join", "2021-07-10 00:00:00", NULL, NULL),
	(7, 10, 9, "l invite", "2021-07-10 00:00:00", NULL, NULL),
	(6, 1, 10, "l join", "2021-07-10 00:00:00", NULL, NULL),
	(6, 10, 10, "l invite", "2021-07-10 00:00:00",NULL, NULL)
;