import React,  { useState } from 'react';
import { getUsers } from './api.js'

function handleSearch(e, settor) {
    settor(e.target.value)
}
function handleKeyDown(event, setText, settor) {
    if (event.key === 'Escape') {
        setText('')
        settor([])
    }
    if (event.key !== 'Enter') {
      return
    }
    if(!event.target.value) {
      return
    }
    getUsers(10, 1, event.target.value)
      .then((resp)=> {
          settor(resp)
      })
      .catch(err => {
        alert(err.toString())
      })
    event.preventDefault();
}

function Header(props) {
    let userauth = props.userauth
    const [searchText, setSearchText] = useState('');
    const [listUsers, setListUsers] = useState([]);
    return (
        <div id="header">
            <span>
                <img src={userauth.avatar} className="icon" title={userauth.fullname}/>
                   [{userauth.username}]</span>
            <input type="text" placeholder="Search User"
                onKeyDown={(e) => handleKeyDown(e, setSearchText, setListUsers)}
                value={searchText}
                onChange={e => handleSearch(e, setSearchText)}/>
            <ul className="search-user">
                {listUsers.map(user => {
                    return <li>{user.fullname}</li>
                })}
            </ul>
        </div>
    );
}

export default Header;
