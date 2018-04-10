create table layer
(
    layer_id serial not null,
    world_id int not null,
    z int not null,
    constraint layer_pkey primary key (layer_id)
);