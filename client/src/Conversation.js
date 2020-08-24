import React from 'react';


function Groups(props) {
    let conversations = Object.values(props.conversations) || []
    return (
        <div id="groups">
            {conversations.map((group, index) => {
                let classname = 'group'
                if (props.selectedGroup.name === group.name) {
                    classname += ' active'
                }
               return <div key={group.id || group.id}
                    className={classname}
                    onClick={e=> props.onClick(group)}>
                    {group.id}
                    {props.subscribed[group.id]?<div className="subscribed"></div>:null}

                </div>
            })}
        </div>
    );
}

export default Groups;
