import React from 'react';
import logo from './logo.svg';
import Logs from './Logs.js'
import Conversation from './Conversation.js'
import wsClient from './ws.js'
import Header from './Header.js'
import {getUser,
  wsHost,
  getUsers,
  getMessages,
  getConversations,
  sendMessage} from './api.js'

function login(username) {
  return getUser(username).then(user => {
    localStorage.setItem("user", JSON.stringify(user))
    return user
  }).catch(err => {
    console.log(err)
    throw err
  })
}
function loadAuthen() {
  try {
    if (localStorage.key("user")) {
      return JSON.parse(localStorage.getItem("user"))
    }
    return null
  } catch (error) {
    return null

  }
}

function preSubscribe(handle, conversations) {
  for(let k in conversations) {
    handle.subscribeGroup(conversations[k])
  }
}

class App extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      userauth: {},
      messages: [],
      wsclient: null,
      inputMessage: '',
      conversations: {
        // 'topic.general': {id: '', name: 'topic.general'},
        // 'topic.room': {id: '', name: 'topic.room'},
      },
      selectedGroup: {},
      subscribed: []
    }
  }
  input = null
  loadConvo(userauth) {
    getConversations(10, 1, userauth.id).then(convo => {
      if(convo.length> 0) {
        let mConvo = {}
        for(let c of convo) {
          mConvo[c.id] = c
        }
        this.setState({conversations: mConvo})
        preSubscribe(this, this.state.conversations)
      }
    }).then(() => {
      let general = 'topic.general'
      if(!Object.keys(this.state.conversations).includes(general)){
        return
      }
      this.handleClickConvo(this.state.conversations[general])
    })
  }
  loadWs() {
    let wsclient = wsClient({
      WebSocket: window.WebSocket,
      url: "ws://" + wsHost + "/ws"
    })
    this.state.wsclient = wsclient

    wsclient.connect(() =>  {
      // let keys = Object.keys(this.state.conversations)
      // this.handleClickConvo(this.state.conversations[keys[0]])
    })

    wsclient.error = (evt) => {
      this.setState(state => ({
        messages: [...state.messages, {text: 'Connection error'}]
      }))
    }

    wsclient.closed = (evt) => {
      this.setState(state => ({
        messages: [...state.messages, {text: 'Connection closed.reconnecting....'}]
      }))
      // reconnect
      wsclient.reconnect()
    }
    wsclient.ondead = () => {
      this.setState(state => ({
        messages: [...state.messages, {text: 'Connection dead'}]
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
  }
  componentDidMount() {
    let userauth = loadAuthen()
    if(!userauth) {
      let username = prompt('Nhập username')
      if (username) {
        login(username).then(user => {
          this.setState({userauth: user})
        })
      }
    }else {
      this.setState({userauth})
    }
    this.loadConvo(userauth)
    this.loadWs()
    window.messages = this.state.messages
    this.scrollToBottom()
  }
  componentWillUnmount() {
    this.state.wsclient.close()
  }
  scrollToBottom = () => {
    this.input.scrollIntoView({ behavior: "smooth" });
  }

  handleKeyDown = (event) => {
    if (event.key !== 'Enter') {
      return
    }
    if(!this.state.inputMessage) {
      return
    }
    sendMessage(this.state.selectedGroup.id, this.state.userauth.id, this.state.inputMessage)
      .then((mes)=> {
        this.setState(state => {
          return {
            messages: [...state.messages, mes]
          }
        }, () => {
          this.scrollToBottom()
        })
      })
      .catch(err => {
        alert(err.toString())
      })
    event.preventDefault();
    this.setState({inputMessage: ''})
  }

  handleOnChange = (event) => {
    this.setState({inputMessage: event.target.value})
  }
  subscribeGroup(group) {
    this.setState(state => ({
      selectedGroup: group,
      subscribed: {...state.subscribed, [group.id]: true}
    }))
  }
  handleClickConvo(convo) {
    if(!this.state.subscribed[convo.id]) {
      this.state.wsclient.subscribe(convo.id)
    }
    getMessages(15, 1, convo.id).then(resp => {
      this.setState({
        messages: (resp.messages || []).reverse()
      })
    })
    this.subscribeGroup(convo)
  }
  render() {
    return (
      <div id="app">
        <Conversation
          conversations={this.state.conversations}
          me={this.state.userauth}
          onClick={(g) => this.handleClickConvo(g)}
          subscribed={this.state.subscribed}
          selectedGroup={this.state.selectedGroup}>
          <Header userauth={this.state.userauth}/>
        </Conversation>

       <div className="leftPanel">
        <Logs messages={this.state.messages}/>

        <form id="form">
          <input type="text" id="myInput" placeholder="Nhập tin nhắn." ref={el => this.input = el}
            onKeyDown={(e) => this.handleKeyDown(e)}
            onChange={e => this.handleOnChange(e)}
            value={this.state.inputMessage}
            />
        </form>
       </div>
      </div>
    )
  }
}

export default App;
