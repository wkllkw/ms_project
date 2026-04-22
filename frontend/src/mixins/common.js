// common-mixin.vue
import {getStore} from "../assets/js/storage";
import config from "../config/config";

export default {
    data() {
        return {}
    },
    methods: {
        routerLink(page, replace = false) {
            const target = typeof page === 'string' ? page : page.path || page.name;
            const current = this.$route.path;
            if (typeof page === 'string' && page === current) return;
            if (replace) {
                this.$router.replace(page).catch(err => {
                    if (err.name !== 'NavigationDuplicated') throw err;
                });
            } else {
                this.$router.push(page).catch(err => {
                    if (err.name !== 'NavigationDuplicated') throw err;
                });
            }
        },
        toHome() {
            const currentOrganization = getStore('currentOrganization', true);
            let home = config.HOME_PAGE;
            if (currentOrganization) {
                home = home + '/' + currentOrganization.code;
            }
            if (this.$route.path !== home) {
                this.$router.push(home);
            }
        },
    }
}
