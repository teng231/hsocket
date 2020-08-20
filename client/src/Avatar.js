
import React from 'react';

const avatarDefault = 'https://st2.depositphotos.com/1104517/11967/v/950/depositphotos_119675554-stock-illustration-male-avatar-profile-picture-vector.jpg'

function Avatar(props) {
  return (
    <div className="avatar">
        <img src={props.avatar || avatarDefault}/>
    </div>
  );
}

export default Avatar;
