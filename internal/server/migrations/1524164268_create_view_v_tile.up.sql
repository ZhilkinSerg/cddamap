create view v_tile as
select 
	w.world_id, 
	l.layer_id, 
	w.name, 
	l.z, 
	w.name || '/o_' || z || '_tiles' as tile_root
from 
	layer l
	inner join world w
		on w.world_id = l.world_id;