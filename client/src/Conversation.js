import React from 'react';
const defaultGroup = "https://geodash.gov.bd/uploaded/people_group/default_group.png"
function render(group, me) {
    if(group.type == 'room') {
        return <span className="convo">
             <img src={defaultGroup}/>
             &nbsp;
             {group.id}</span>
    }

    if(group.type == 'chat') {
        let members = group.members || {}
        let idFriend = Object.keys(members).find(k => k != me.id)
        if(!idFriend) {
            return <span className="convo">
                <img src={defaultGroup}/>
                &nbsp;
                {group.id}
            </span>
        }
        return <span className="convo">
            <img src={members[idFriend].avatar}/>
            &nbsp;
            {members[idFriend].fullname}
        </span>
    }
}
function Groups(props) {
    let conversations = Object.values(props.conversations) || []
    debugger
    return (
        <div id="groups">
            {props.children}
            {conversations.map((group, index) => {
                let classname = 'group'
                if (props.selectedGroup.id === group.id) {
                    classname += ' active'
                }
               return <div key={group.id || group.id}
                    className={classname}
                    onClick={e=> props.onClick(group)}>
                    {render(group, props.me)}
                    {props.subscribed[group.id]?<div className="subscribed"></div>:null}
                </div>
            })}

        </div>
    );
}

export default Groups;
