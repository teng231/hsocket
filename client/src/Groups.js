import React from 'react';


function Groups(props) {
    let groups = Object.values(props.groups) || []
    console.log(1111, props.selectedGroup.name)
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
                    `(${group.id}): ${group.name}`
                </div>
            })}
        </div>
    );
}

export default Groups;
