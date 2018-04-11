create table layer
(
    layer_id serial not null,
    world_id int not null,
    z int not null,
    created_at timestamp with time zone not null default now(),    
    constraint layer_pkey primary key (layer_id)
);

alter table layer add constraint fk_layer_world foreign key(world_id) references world(world_id);
