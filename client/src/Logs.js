import React from 'react';
import Avatar from './Avatar.js'

const elementDisplay = (message) => {
  if(message.notification_type) {
    return <p className="notification">{message.raw}</p>
  }
  return <span>
    <Avatar avatar={message.avatar}/>
    [{message.send_to}]- {message.sender}: {message.raw}
  </span>
}
function Log(props) {
  return (
    <div id="log">
        {(props.messages || []).map((message, index) => {
            return <div key={message.id || index}>

              {elementDisplay(message)}
            </div>
        })}
    </div>
  );
}

export default Log;
