// Used in multiple functions and is constant
var roomCode = $("#room-code").val();

$("#copy-shareable-link").click(copyShareableLink);

function copyShareableLink() {
	$("#shareable-link-content").focus().select();
	document.execCommand("copy");
}

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
		if(item) {
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
	}

	$(".add").click(addToQueue);

}

function appendToSongList(item) {
	var elem = $("<div class='queue-song' id=" + item.id + " style='display: none'>" +
					"<img src=" + item.album.images[2].url + ">" +
					"<div class='details'>" +
						"<h3>" + item.name + "</h3>" +
						"<p>" + item.artists[0].name + "</p>" +
					"</div>" +
				  "</div>"
				);

	$("#queue-songs-container").append(elem);
	elem.show('slow');
}

function removeSongFromSongList(trackID) {
	var elem = $("#" + trackID)
	elem.slideUp('slow', function() {
		elem.remove();
	});
}

function addToQueue(e) {
	var songID = e.target.id

	$.ajax({
		type: "POST",
		url: "/add",
		data: { "songID": songID, "roomCode": roomCode },
		success: function() {
			e.target.innerHTML = "Added!"
			e.target.style.backgroundColor = '#1DB954';
		},
		error: function(res) {
			var j = JSON.parse(res.responseText)
			var elem = $("<p class='popup-error'>" + j.msg + "</p>");	
			$(".content").append(elem)
			elem.show().delay(2000).fadeOut();
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
		success: function(res) {
			// Update the room code for confirmStart
			j = JSON.parse(res);
			roomCode = j.roomCode;
			$("#start-modal").fadeIn();
		}
	});
}

$("#close-room").click(closeRoom);

function closeRoom() {
	$.ajax({
		type: "POST",
		url: "/room/close",
		data: { "roomCode": roomCode },
		success: function() {
			location.reload();
		}
	});
}

$("#confirm-start").click(confirmStart);

function confirmStart() {
	$.ajax({
		type: "POST",
		url: "/room/start",
		data: { "roomCode": roomCode },
		success: function() {
			location.reload();
		}
	});
}

$("#open-search-modal").click(openSearchModal);

function openSearchModal() {
	$("#search-modal").show(500);
	$(".search-bar").focus();
}

$("#close-modal").click(closeSearchModal);

function closeSearchModal() {	
	$("#search-modal").hide(500);
}

$("#veto-song").click(vetoSong);

function vetoSong() {
	$.ajax({
		type: "POST",
		url: "/room/veto",
		data: { "roomCode": roomCode },
		success: function() {
			$("#veto-song").html("Voted!");

			// Don't allow the button to be clicked again
			$("#veto-song").off('click');
		}
	});
}

function updateVetoCount(count) {
	$("#veto-count").text(count + " ");
}


function updateUserCount(count) {
	$("#user-count").text(" " + count);
}

function resetVoteBtn() {
	$("#veto-song").html("Veto Current Song");

	// Add click handler back
	$("#veto-song").on('click', vetoSong);
}
