// Used in multiple functions and is constant
var roomCode = $("#room-code").val();

$("#search-form").submit(searchSubmit);

function searchSubmit() {
	var songName = $(".search-bar").val()
	
	$.ajax({
		type: "POST",
		url: "/search",
		data: { "songName": songName, "roomCode": roomCode },
		success: updateSongList
	})

	// Don't refresh the page
	return false;
}

function updateSongList(resultData) {
	// Clear out any songs added earlier
	$("#song-container").empty()

	for(var i = 0; i < 10; i++) {
		var item = resultData[i]
		var elem = $("<div class='queue-song'>" +
						"<img src=" + item.album.images[2].url + ">" +
						"<div class='details'>" +
							"<h3>" + item.name + "</h3>" +
							"<p>" + item.artists[0].name + "</p>" +
						"</div>" +
						"<a class='cta-btn add' id=" + item.id + ">" + "Add </a>" +
					  "</div>"
					);

		$("#song-container").append(elem);
	}

	$(".add").click(addToQueue);

}

function appendToSongList(item) {
	var elem = $("<div class='queue-song'>" +
					"<img src=" + item.album.images[2].url + ">" +
					"<div class='details'>" +
						"<h3>" + item.name + "</h3>" +
						"<p>" + item.artists[0].name + "</p>" +
					"</div>" +
				  "</div>"
				);

	$("#queue-songs-container").append(elem);
}

function addToQueue(e) {
	var songID = e.target.id

	$.ajax({
		type: "POST",
		url: "/add",
		data: { "songID": songID, "roomCode": roomCode },
		success: function() {
			console.log(e)
			e.target.innerHTML = "Added!"
			e.target.style.backgroundColor = '#1DB954';
		}
	});
}

$("#create-playlist").click(createPlaylist);

function createPlaylist() {
	$.ajax({
		type: "POST",
		url: "/playlist/create",
		success: function() {
			location.reload();
		}
	});
}

$("#open-room").click(openRoom);

function openRoom() {
	$.ajax({
		type: "POST",
		url: "/room/open",
		success: function() {
			location.reload();
		}
	});
}

$("#close-room").click(closeRoom);

function closeRoom() {
	$.ajax({
		type: "POST",
		url: "/room/close",
		success: function() {
			location.reload();
		}
	});
}

$("#open-search-modal").click(openSearchModal);

function openSearchModal() {
	$(".modal-bg").show(500);
	$("#search-modal").show(500);
}

// Catch click outside of the modal
$(".modal-bg").click(closeSearchModal);

function closeSearchModal() {	
	$(".modal-bg").hide(500);
	$("#search-modal").hide(500);
}

