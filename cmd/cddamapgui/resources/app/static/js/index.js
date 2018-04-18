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
        var zl = 8;

		var map = L.map('map', {
			crs: L.CRS.Simple,
			//maxZoom: 8
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
			response = index.cell(1, p.x, p.y)
			console.log(response)
			selectedCell.clearLayers()
			selectedCell.addData(response)
		});
	},
	cell: function(l, x, y) {
        let message = {"name": "cell"};
		message.payload = {L: l, X: x, Y: y}

        asticode.loader.show();
        astilectron.sendMessage(message, function(message) {
            asticode.loader.hide();
            if (message.name === "error") {
                asticode.notifier.error(message.payload);
                return
			}
			return message.payload
        })
    },
};