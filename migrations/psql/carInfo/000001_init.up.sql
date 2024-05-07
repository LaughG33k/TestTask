CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table if not exists cars (
    id bigserial primary key,
    reg_num varchar(100) unique not null,
    mark varchar(100),
    model varchar(100),
    year integer,
    
    owner_name varchar(100),
    owner_surname varchar(100),
    owner_patranomic varchar(100)
);

create index if not exists reg_num on cars (reg_num);

