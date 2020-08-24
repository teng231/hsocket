import React from 'react';
import Avatar from './Avatar.js'

const elementDisplay = (message) => {
  if(message.notification_type) {
    return <p className="notification">{message.text}</p>
  }
  return <div>
    <Avatar avatar={message.avatar}/>
    <span className="message" >{message.text}</span>
  </div>
}
function Log(props) {
  return (
    <div id="log">
        {(props.messages || []).map((message, index) => {
            return <div key={message.id || index}  className="border">
              {elementDisplay(message)}
            </div>
        })}
    </div>
  );
}

export default Log;
