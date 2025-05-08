CREATE TABLE blog (
	blogid uuid,
	userid uuid,
	title varchar,
    content varchar,
	releasetime timestamp DEFAULT NOW(),
	primary key (blogid)
);

CREATE TABLE users (
	id uuid,
	username VARCHAR(30),
	password VARCHAR,
	refreshToken VARCHAR,
	admin BOOLEAN,
	primary key (id)
);