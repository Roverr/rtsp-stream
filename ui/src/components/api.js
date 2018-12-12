
import axios from 'axios';

export class APIHandler{
    constructor(url) {
        this.url = url;
    }
    listStreams() {
        return axios.get(`${this.url}/list`);
    }
}