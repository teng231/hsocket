const wsHost = 'localhost:3001'

function sendMessage(convoid, sender_id, body) {
	if(!body) {
		return
	}
	return fetch('http://' + wsHost + '/ws-firer', {
		method: 'POST', // *GET, POST, PUT, DELETE, etc.
		mode: 'cors', // no-cors, *cors, same-origin
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({
			conversation_id :convoid,
			text: body,
			encoding: "text/plain",
            sender_id: sender_id,
            type: "raw",
			// conn_id: _ws.conn.conn_id
		}) // body data type must match "Content-Type" header
	}).then(rs => rs.json())
}

function getMessages(limit, page, convoid) {
	return fetch('http://' + wsHost + '/messages/' + convoid + "?limit=" + limit + "&page="+page, {
		method: 'GET', // *GET, POST, PUT, DELETE, etc.
		mode: 'cors', // no-cors, *cors, same-origin
		headers: {
			'Content-Type': 'application/json'
		},
	}).then(rs => rs.json())
}

function getUsers(limit, page, username) {
	return fetch('http://' + wsHost + '/users?limit=' + limit + "&page="+page + "&username=" + username, {
		method: 'GET', // *GET, POST, PUT, DELETE, etc.
		mode: 'cors', // no-cors, *cors, same-origin
		headers: {
			'Content-Type': 'application/json'
		},
	}).then(rs => rs.json())
}

function getUser(username) {
	return fetch('http://' + wsHost + '/users/'+ username, {
		method: 'GET', // *GET, POST, PUT, DELETE, etc.
		mode: 'cors', // no-cors, *cors, same-origin
		headers: {
			'Content-Type': 'application/json'
		},
	}).then(rs => rs.json())
}
function getConversations(limit, page, userid) {
	return fetch('http://' + wsHost + '/conversations?user_id=' + userid + "&limit=" + limit + "&page=" + page, {
		method: 'GET', // *GET, POST, PUT, DELETE, etc.
		mode: 'cors', // no-cors, *cors, same-origin
		headers: {
			'Content-Type': 'application/json'
		},
	}).then(rs => rs.json())
}

export {
    wsHost,
    getUser,
    getUsers,
    getMessages,
    sendMessage,
    getConversations
}
