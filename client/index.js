const url = 'ws://localhost:8000/'
const connection = new WebSocket(url)

connection.onopen = () => {
    console.log("Connected...")
    connection.send('Message From Client')
}

connection.onerror = (error) => {
    console.log(`WebSocket error: ${error}`)
}

connection.onmessage = (e) => {
    console.log(e.data)
}