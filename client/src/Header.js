import React from 'react';


function Header(props) {
    let userauth = props.userauth
    return (
        <div id="header">
            <span>
                <img src={userauth.avatar} className="icon" title={userauth.fullname}/>
                   [{userauth.username}]</span>
            <input type="text" placeholder="Search User"/>
        </div>
    );
}

export default Header;
