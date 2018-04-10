create table cell
(
    cell_id serial not null,
    layer_id int not null,
    name character varying, 
    the_geog geography(POINT,4326) not null,
    constraint cell_pkey primary key (cell_id)
);

alter table cell add constraint fk_cell_layer foreign key(layer_id) references layer(layer_id);