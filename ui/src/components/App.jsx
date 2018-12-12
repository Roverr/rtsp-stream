import React, { Component } from 'react';
import { Row, Button, Container, InputGroupAddon, InputGroup, Input } from 'reactstrap';
import ReactHls from 'react-hls';
import { APIHandler,  } from './api';

class App extends Component {
    constructor(props) {
        super(props);
        this.url = 'http://localhost:8080';
        this.apiHandler = new APIHandler(this.url);
        this.state = { streams: [], current: '' };
    }
    getStream() {
        if (this.state.current === '') {
            return null;
        }
        return (
            <ReactHls url={this.state.current} />
        );
    }
    render() {
        return (
          <Container style={{marginTop: '2em'}}>
            <Row>
            <InputGroup>
                <InputGroupAddon addonType="prepend"><Button> Add URI </Button></InputGroupAddon>
                <Input placeholder="http://username:password@host:port/subroute" />
            </InputGroup>
            </Row>
            <Row>
                Available streams: {this.state.streams}
            </Row>
            {this.getStream()}
          </Container>
        );
    }
    componentDidMount() {
        this.apiHandler.listStreams().then((resp) => {
            if (resp.data.length > 0) {
                this.state = {
                    ...this.state,
                    streams: resp.data,
                };
            }
        }).catch((err) => console.log(err));
    }
}


export default App;
