import StorageObject from 'ember-local-storage/session/object';

const Storage = StorageObject.extend();

// Uncomment if you would like to set initialState
Storage.reopenClass({
  initialState() {
    return {id: ''};
  }
});

export default Storage;
