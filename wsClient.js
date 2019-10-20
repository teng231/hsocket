
var _conn = null
var _topics = []


function wsClient(configs){
	let me = {}
	me.onmessage = function() {}

	me.connect = function(cb) {
		_conn = new configs.WebSocket(configs.url)
		_conn.onopen = function() {
			if(typeof cb === 'function') cb()
		}
		_conn.onmessage = function(evt) {
			let data = evt.data
			if(!data) cb(Error('message empty'), null, evt)
			try{
				let jsonData = JSON.parse(data)
				me.onmessage(null, jsonData, evt)
			} catch(err) {
				me.onmessage(err, null, evt)
			}
		}
		_conn.onerror = function(evt) {
			me.onerror(evt)
		}
	}

	me.subscribe = function(topic) {
		if(_topics.includes(topic)) return
		_conn.send(JSON.stringify({
			Type: 'subscribe',
			topic: topic
		}))
		_topics.push(topic)
	}

	me.unsubscribe = function(topic) {
		if(!_topics.includes(topic)) return
		_conn.send(JSON.stringify({
			Type: 'unsubscribe',
			topic: topic
		}))
		_topics = _topics.filter(tp => tp == topic)
	}

	me.send = function(topic, payload) {
		if(!_topics.includes(topic)) {
			me.subscribe(topic)
		}
		var body = ''
		if(typeof payload == 'string') {
			body = payload
		}
		if(typeof payload == 'object') {
			body = JSON.stringify(payload)
		}
		_conn.send(JSON.stringify({
			topic: topic,
			body: body
		}))
	}

	me.getTopics = function() {
		return _topics
	}

	me.destroy = function() {
		if(!_conn) return
		_conn.close()
		_topics = []
	}

	me.reconnect = function(cb) {
		var topics = me.getTopics()
		me.destroy()
		debugger
		me.connect(function() {
			topics.forEach(tp => {
				me.subscribe(tp)
			})
			if(typeof cb == 'function') cb()
		})
	}

	return me
}
