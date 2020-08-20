let username = ''

function wsClient(configs){
	var MAX_RETRY = 7
	var CURRENT_RETRIED = 0
	var _ws = {
		conn: null,
		backoff: 1000,
		topics: {},
		url: configs.url,
	}
	let me = {}
	me.onmessage = function() {}

	me.sendEvent = function(topic, notitype) {
		return fetch('http://' + document.location.host + '/ws-firer', {
			method: 'POST', // *GET, POST, PUT, DELETE, etc.
			mode: 'cors', // no-cors, *cors, same-origin
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				send_to:topic,
				notification_type: notitype,
				to_group: true,
				encoding: "text/plain",
				sender: username,
				conn_id: _ws.conn.conn_id,
			}) // body data type must match "Content-Type" header
		}).then(rs => rs.json())
	}

	me.connect = function(cb) {
		_ws.conn = new configs.WebSocket(_ws.url)
		_ws.conn.id = CURRENT_RETRIED
		_ws.conn.onopen = function() {
			if(typeof cb === 'function') cb()
			CURRENT_RETRIED = 0
		}
		_ws.conn.onclose = function(evt) {
			if(CURRENT_RETRIED > MAX_RETRY) {
				me.ondead()
			}
			setTimeout(() => {
				me.closed(evt)
			}, CURRENT_RETRIED * _ws.backoff)
		}
		_ws.conn.onmessage = function(evt) {
			let data = evt.data
			if(!data) cb(Error('message empty'), null, evt)
			try{
				let jsonData = JSON.parse(data)
				// console.log(data)
				if(data.notification_type === "connected") {
					console.log('connection id:', data.conn_id)
					window.conn_id = data.conn_id
					return
				}
				me.onmessage(null, jsonData, evt)
			} catch(err) {
				me.onmessage(err, null, evt)
			}
		}
	}

	me.subscribe = function(topic) {
		_ws.conn.send(JSON.stringify({
			notification_type: 'subscribe',
			send_to: topic,
			to_group: true,
			sender: username,
		}))
		// me.sendEvent(topic, 'subscribe')
		_ws.topics[topic] = true
	}

	me.unsubscribe = function(topic) {
		_ws.conn.send(JSON.stringify({
			notification_type: 'unsubscribe',
			send_to: topic,
			to_group: true,
			sender: username,
		}))
		// me.sendEvent(topic, 'unsubscribe')
		delete _ws.topics[topic]
	}

	me.send = function(topic, payload) {
		if(!_ws.topics.includes(topic)) {
			me.subscribe(topic)
		}
		var body = ''
		if(typeof payload === 'string') {
			body = payload
		}
		if(typeof payload === 'object') {
			body = JSON.stringify(payload)
		}
		_ws.conn.send(JSON.stringify({
			send_to: topic,
			sender: username,
			encoding: "text/plain",
			raw: body
		}))
	}

	me.getTopics = function() {
		return Object.keys(_ws.topics)
	}

	me.destroy = function() {
		if(!_ws.conn) return
		_ws.conn.close()
		_ws.topics = {}
	}

	me.reconnect = function(cb) {
		if(CURRENT_RETRIED > MAX_RETRY) return
		var topics = me.getTopics()
		if([_ws.conn.CONNECTING, _ws.conn.OPEN].includes(_ws.conn.readyState)) return
		CURRENT_RETRIED++
		me.connect(function() {
			topics.forEach(tp => me.subscribe(tp))
			if(typeof cb === 'function') cb()
		})
	}

	return me
}
export default wsClient