const Component: React.FunctionComponent<{}> = () => {
    let name = 'id'
    return (
        <div>
            <h1 id={name}>My Component</h1>
            {[1, 2, 3].map(item => (
                <p key={item}>{item}</p>
            ))}
        </div>
    )
}
