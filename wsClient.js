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
				me.onmessage(null, jsonData, evt)
			} catch(err) {
				me.onmessage(err, null, evt)
			}
		}
	}

	me.subscribe = function(topic) {
		_ws.conn.send(JSON.stringify({
			Type: 'subscribe',
			topic: topic
		}))
		_ws.topics[topic] = true
	}

	me.unsubscribe = function(topic) {
		_ws.conn.send(JSON.stringify({
			Type: 'unsubscribe',
			topic: topic
		}))
		delete _ws.topics[topic]
	}

	me.send = function(topic, payload) {
		if(!_ws.topics.includes(topic)) {
			me.subscribe(topic)
		}
		var body = ''
		if(typeof payload == 'string') {
			body = payload
		}
		if(typeof payload == 'object') {
			body = JSON.stringify(payload)
		}
		_ws.conn.send(JSON.stringify({
			topic: topic,
			body: body
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
			if(typeof cb == 'function') cb()
		})
	}

	return me
}
