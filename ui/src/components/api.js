
import axios from 'axios';

export class APIHandler{
    constructor(url) {
        this.url = url;
    }
    getUrl() {
        return this.url;
    }
    listStreams() {
        return axios.get(`${this.url}/list`).then((res) => {
            return res.data.map(
                ({ path }) => `${this.url}/stream/${path}/index.m3u8`,
            );
        });
    }
    startStream(uri) {
        return axios.post(`${this.url}/start`, {uri});
    }
}