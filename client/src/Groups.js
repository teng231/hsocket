import React from 'react';


function Groups(props) {
    let groups = Object.values(props.groups) || []
    return (
        <div id="groups">
            {groups.map((group, index) => {
                let classname = 'group'
                if (props.selectedGroup.name === group.name) {
                    classname += ' active'
                }
               return <div key={group.id || group.name}
                    className={classname}
                    onClick={e=> props.onClick(group)}>
                    {group.name}
                    {props.subscribed[group.name]?<div className="subscribed"></div>:null}

                </div>
            })}
        </div>
    );
}

export default Groups;
