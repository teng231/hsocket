import React from 'react';

function Log(props) {
  return (
    <div id="log">
        {(props.messages || []).map((message, index) => {
            return <div key={message.id || index}>
                `(${message.send_to})${message.sender}: ${message.raw}`</div>
        })}
    </div>
  );
}

export default Log;
