
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
                ({ uri }) => `${this.url}${uri}`,
            );
        });
    }
    startStream(uri) {
        return axios.post(`${this.url}/start`, {uri});
    }
}
