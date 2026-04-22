<script>
import ProjectList from '../list/index';

export default {
    name: 'ProjectRecycleIndex',
    extends: ProjectList,
    data() {
        return {
            selectBy: 'deleted',
        };
    },
};
</script>
