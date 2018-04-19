let index = {
    init: function() {
        asticode.loader.init();
        asticode.modaler.init();
        asticode.notifier.init();

        document.addEventListener('astilectron-ready', function() {
            index.mapstuff();
        })
    },
    mapstuff: function() {
        var zl = 5;

		var map = L.map('map', {
			crs: L.CRS.Simple,
		}).setView([-128, 128], 1);

		var tiles = L.tileLayer('http://localhost:8080/{z}/{x}/{y}.png', {
			maxZoom: 10,
			maxNativeZoom: zl,
			attribution: '',
			noWrap: true,
			crs: L.CRS.Simple
		}).addTo(map);

		var selectedCell = L.geoJSON(null, {
			coordsToLatLng: function (nc) {
				return map.unproject([nc[0], nc[1]], zl);
			}
		}).bindTooltip(function (layer) {
			return layer.feature.properties.name;
		}).addTo(map);

		map.on('click', function (e) {
			p = map.project(e.latlng, zl)

			let message = {"name": "cell"};
			message.payload = {L: 10, X: p.x, Y: p.y}

			//asticode.loader.show();
			astilectron.sendMessage(message, function(message) {
				//asticode.loader.hide();
				if (message.name === "error") {
					asticode.notifier.error(message.payload);
					return
				}
				selectedCell.clearLayers()
				selectedCell.addData(message.payload)
			})
		});
	}
};