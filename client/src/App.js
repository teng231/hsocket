import React from 'react';
import logo from './logo.svg';
import Logs from './Logs.js'
import Groups from './Groups.js'
import wsClient from './ws.js'
import Header from './Header.js'
const wsHost = 'localhost:3001'

function postData(topic, username, body) {
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
			conversation_id :topic,
			text: body,
			encoding: "text/plain",
			sender: username,
			// conn_id: _ws.conn.conn_id
		}) // body data type must match "Content-Type" header
	}).then(rs => rs.json())
}

function getGroups() {
	return fetch('http://' + wsHost + '/topics', {
		method: 'GET', // *GET, POST, PUT, DELETE, etc.
		mode: 'cors', // no-cors, *cors, same-origin
		headers: {
			'Content-Type': 'application/json'
		},
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





class App extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      messages: [],
      wsclient: null,
      inputMessage: '',
      username: 'teng.' + Date.now(),
      groups: {
        'topic.general': {id: '', name: 'topic.general'},
        'topic.x': {id: '', name: 'topic.x'},
        'chat.x': {id: '', name: 'chat.x'}
      },
      selectedGroup: {},
      subscribed: []
    }
  }

  componentDidMount() {
    let username = prompt('Nhập username')
    if (username) {
      this.setState({username})
    }

    let wsclient = wsClient({
      WebSocket: window.WebSocket,
      url: "ws://" + wsHost + "/ws"
    })
    this.state.wsclient = wsclient

    wsclient.connect(() =>  {
      let defaultGroup = 'topic.general'
      wsclient.subscribe(this.state.groups[defaultGroup].name)
      this.subscribeGroup(this.state.groups[defaultGroup])
      getMessages(10, 1, 'topic.general').then(resp => {
        console.log(resp.messages)
        this.setState({
          messages: resp.messages || []
        })
      })
    })

    wsclient.error = (evt) => {
      this.setState((state, props) => ({
        messages: [...state.messages, {raw: 'Connection error'}]
      }))
    }

    wsclient.closed = (evt) => {
      this.setState((state, props) => ({
        messages: [...state.messages, {raw: 'Connection closed.reconnecting....'}]
      }))
      // reconnect
      wsclient.reconnect()
    }
    wsclient.ondead = () => {
      this.setState((state, props) => ({
        messages: [...state.messages, {raw: 'Connection dead'}]
      }))
    }

    wsclient.onmessage = (err, message, evt) => {
      console.log(message)

      if(message.notification_type === 'connected') {
        return
      }

      this.setState((state, props) => ({
        messages: [...state.messages,message]
      }))
    }
    getGroups().then(groups => {
      console.log(groups)
      this.setState(state => ({groups: {...state.groups, ...groups}}))
    })
    setInterval(() => {
      getGroups().then(groups => {
        console.log(groups)
        this.setState(state => ({groups: {...state.groups, ...groups}}))
      })
    }, 10 * 1000)
    window.messages = this.state.messages
  }
  componentWillUnmount() {
    this.state.wsclient.close()
  }
  handleKeyDown = (event) => {
    if (event.key !== 'Enter') {
      return
    }
    postData(this.state.selectedGroup.name, this.state.username, this.state.inputMessage)
    event.preventDefault();
    this.setState({inputMessage: ''})
  }
  handleOnChange = (event) => {
    this.setState({inputMessage: event.target.value})
  }
  subscribeGroup(group) {
    this.setState(state => ({
      selectedGroup: group,
      subscribed: {...state.subscribed, [group.name]: true}
    }))
  }
  handleClickGroup(group) {
    if(!this.state.subscribed[group.name]) {
      this.state.wsclient.subscribe(group.name)
    }
    getMessages(10, 1, group.name).then(resp => {
      this.setState({
        messages: resp.messages || []
      })
    })
    this.subscribeGroup(group)
  }
  render() {
    return (
      <div id="app">

        <Groups groups={this.state.groups}
          onClick={(g) => this.handleClickGroup(g)}
          subscribed={this.state.subscribed}
          selectedGroup={this.state.selectedGroup}/>

        <Logs messages={this.state.messages}/>
        <Header />
        <form id="form">
          <input type="text" id="myInput" placeholder="Nhập tin nhắn."
            onKeyDown={(e) => this.handleKeyDown(e)}
            onChange={e => this.handleOnChange(e)}
            value={this.state.inputMessage}
            />
        </form>
      </div>
    )
  }
}

export default App;