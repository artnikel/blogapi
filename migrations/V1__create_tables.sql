CREATE TABLE blog (
	blogid uuid,
	profileid uuid,
	title varchar,
    content varchar,
	operationtime timestamp DEFAULT NOW(),
	primary key (blogid)
);