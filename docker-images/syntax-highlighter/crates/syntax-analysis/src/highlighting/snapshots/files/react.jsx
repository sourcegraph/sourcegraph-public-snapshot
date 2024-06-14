const ComponentJsx = () => {
    let name = 'id'
    return (
        <div>
            <p />
            <h1 id={name}>My Component</h1>
            {[1, 2, 3].map(item => (
                <p key={item}>{item}</p>
            ))}
        </div>
    )
}
