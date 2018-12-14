import React, { Component } from 'react';
import ReactHls from 'react-hls';
import {
    Row,
    Button,
    Container,
    InputGroupAddon,
    InputGroup,
    Input,
    Col,
    FormGroup,
    Card
} from 'reactstrap';

import { APIHandler } from './api';

class App extends Component {
    constructor(props) {
        super(props);
        this.state = { streams: [], current: null };
        /** @property {APIHandler} apiHandler */
        this.apiHandler = this.props.apiHandler || new APIHandler(process.env.API_URL || 'http://localhost:8080');
        /** @property {HTMLInputElement} uriInput */
        this.uriInput;
    }

    componentDidMount() {
        this.apiHandler.listStreams().then(
            (streams) => this.setState({
                streams,
                current: streams.length ? 0 : null,
            })
        ).catch((e) => console.log(e));
    }

    addInputStream() {
        this.apiHandler.startStream(this.uriInput.value).then((res) => {
            const { uri } = res.data;
            this.setState(({ streams, current }) => ({
                streams: [ ...streams, `${this.apiHandler.getUrl()}${uri}` ],
                current: current || 0,
            }));
        }).catch((e) => console.log(e));
    }

    render() {
        return (
          <Container className="my-4">
            <Row>
                <Col md={{size: 8, offset: 2}}>
                    {this.renderPlayer()}
                </Col>
            </Row>
            <Row>
                <Col md={{size: 8, offset: 2}}>
                    {this.renderInput()}
                    {this.renderList()}
                </Col>
            </Row>
          </Container>
        );
    }

    renderInput() {
        return (
            <InputGroup className="my-1">
                <Input
                    innerRef={(elem) => (this.uriInput = elem)}
                    placeholder="rtsp://username:password@host:port/subroute"
                />
                <InputGroupAddon addonType="append">
                    <Button color="primary" onClick={this.addInputStream.bind(this)}>Add</Button>
                </InputGroupAddon>
            </InputGroup>
        );
    }

    renderList() {
        const onChange = (ev) => this.setState({ current: parseInt(ev.target.value, 10) })
        const playStreamFactory = (current) => () => this.setState({ current });
        const options = this.state.streams.map(
            (uri, offset) => (
                <option key={uri} value={offset} onClick={playStreamFactory(offset)}>
                    {uri}
                </option>
            ),
        );
        return (
            <Input type="select" onChange={onChange} className="my-1">
                {options}
            </Input>
        )
    }

    renderPlayer() {
        const props = { style: {textAlign: "center", position: "relative"} };
        const content = this.state.current === null
            ? <span className="display-5 py-4">Select or add a stream below.</span>
            : <ReactHls width="100%" url={`${this.state.streams[this.state.current]}`} autoplay />;
        return (
            <FormGroup>
                <Card {...props}>{content}</Card>
            </FormGroup>
        );
    }
}

export default App;
