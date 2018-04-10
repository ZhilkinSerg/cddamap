create table world
(
    world_id serial not null,
    name character varying not null, 
    constraint world_pkey primary key (world_id)
);