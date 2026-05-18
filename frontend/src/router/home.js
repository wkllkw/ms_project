/**
 * Home
 */
export default [
    {
        name: 'home',
        path: '/home/:code',
        component: resolve => require(['@/views/home/index'], resolve),
        meta: {model: 1, info: {id: 'home', title: '首页', fullUrl: 'home', show_slider: true, is_inner: false}, permission: 'home'},
    },
    {
        //项目概览（首页）
        name: 'projectIndex',
        path: '/project/space/index/:code',
        component: resolve => require(['@/views/project/space/index'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
    },
    {
        //任务看板
        name: 'task',
        path: '/project/space/task/:code',
        component: resolve => require(['@/views/project/space/task'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
        children: [
            {
                //任务详情
                name: 'taskdetail',
                path: 'detail/:taskCode',
                component: resolve => require(['@/views/project/space/taskdetail'], resolve),
                meta: {model: 10, info: {show_slider: false}, permission: 'project.list'},
            },
        ]
    },
    {
        name: 'projectOverview',
        path: '/project/space/overview/:code',
        component: resolve => require(['@/views/project/space/overview'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
    },
    {
        name: 'projectFiles',
        path: '/project/space/files/:code',
        component: resolve => require(['@/views/project/space/files'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
    },
    {
        name: 'projectFeatures',
        path: '/project/space/features/:code',
        component: resolve => require(['@/views/project/space/features'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
    },
    {
        name: 'projectEvents',
        path: '/project/space/events/:code',
        component: resolve => require(['@/views/project/space/events'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
    },
    {
        name: 'projectGantt',
        path: '/project/space/gantt/:code',
        component: resolve => require(['@/views/project/space/gantt'], resolve),
        meta: {model: 10, info: {show_slider: false, is_inner: true}, permission: 'project.list'},
    },
    {
        name: 'projectListDefault',
        path: '/project/list',
        redirect: '/project/list/my',
        meta: {model: 10, info: {show_slider: true, is_inner: false}, permission: 'project.list'},
    },
    {
        name: 'projectList',
        path: '/project/list/:type',
        component: resolve => require(['@/views/project/list/index'], resolve),
        meta: {model: 10, info: {show_slider: true, is_inner: false}, permission: 'project.list'},
    },
    {
        name: 'projectTemplate',
        path: '/project/template',
        component: resolve => require(['@/views/project/template/index'], resolve),
        meta: {model: 10, info: {show_slider: true, is_inner: false}, permission: 'project.template'},
    },
    {
        name: 'projectTemplateTaskStages',
        path: '/project/template/taskstages/:code',
        component: resolve => require(['@/views/project/template/taskStages'], resolve),
        meta: {model: 10, info: {show_slider: true, is_inner: false}, permission: 'project.template'},
    },

    {
        name: 'projectAnalysis',
        path: '/project/analysis',
        component: resolve => require(['@/views/project/analysis/index'], resolve),
        meta: {model: 10, info: {show_slider: true, is_inner: false}, permission: 'project.analysis'},
    },
    {
        //邀请链接
        name: 'inviteFromLink',
        path: '/invite_from_link/:code',
        component: resolve => require(['@/views/common/inviteFromLink'], resolve),
        meta: {model: 0, info: {show_slider: false}},
    },
    {
        name: 'calendar',
        path: '/calendar',
        component: resolve => require(['@/views/common/calendar'], resolve),
        meta: {model: 25, info: {show_slider: true, is_inner: false}, permission: 'calendar'},
    },
    {
        name: 'accountSettingBase',
        path: '/account/setting/base',
        component: resolve => require(['@/views/account/setting/base'], resolve),
        meta: {model: 0, info: {show_slider: false, is_inner: true}},
    },
    {
        name: 'accountSettingSecurity',
        path: '/account/setting/security',
        component: resolve => require(['@/views/account/setting/security'], resolve),
        meta: {model: 0, info: {show_slider: false, is_inner: true}},
    },
    {
        name: 'members',
        path: '/members',
        component: resolve => require(['@/views/members/index'], resolve),
        meta: {model: 30, info: {show_slider: true, is_inner: false}, permission: 'members.index'},
    },
    {
        name: 'memberProfile',
        path: '/members/profile/:code',
        component: resolve => require(['@/views/members/profile'], resolve),
        meta: {model: 30, info: {show_slider: true, is_inner: false}, permission: 'members.index'},
    },
    {
        name: 'organizationIndex',
        path: '/organization/index',
        component: resolve => require(['@/views/organization/index'], resolve),
        meta: {model: 30, info: {show_slider: true, is_inner: false}, permission: 'project.manage'},
    },
    {
        name: 'organizationMembers',
        path: '/organization/members',
        component: resolve => require(['@/views/organization/members/index'], resolve),
        meta: {model: 30, info: {show_slider: true, is_inner: false}, permission: 'organization.member'},
    },
    {
        name: 'notifyNotice',
        path: '/notify/notice',
        component: resolve => require(['@/views/notify/notice'], resolve),
        meta: {model: 20, info: {show_slider: true, is_inner: false}, permission: 'notify.notice'},
    },
    {
        name: 'notifySystem',
        path: '/notify/system',
        component: resolve => require(['@/views/notify/notice'], resolve),
        meta: {model: 20, info: {show_slider: true, is_inner: false}, permission: 'notify.system'},
    },
    {
        name: 'systemAccount',
        path: '/system/account',
        component: resolve => require(['@/views/system/account/index'], resolve),
        meta: {model: 40, info: {show_slider: true, is_inner: false}, permission: 'system.account'},
    },
    {
        name: 'systemAccountAuth',
        path: '/system/account/auth',
        component: resolve => require(['@/views/system/account/auth'], resolve),
        meta: {model: 40, info: {show_slider: true, is_inner: false}, permission: 'system.account.auth'},
    },
    {
        name: 'systemAccountApply',
        path: '/system/account/apply/:id',
        component: resolve => require(['@/views/system/account/apply'], resolve),
        meta: {model: 40, info: {show_slider: true, is_inner: false}, permission: 'system.account.auth'},
    },
    {
        name: 'systemConfigMenu',
        path: '/system/config/menu',
        component: resolve => require(['@/views/system/config/menu'], resolve),
        meta: {model: 40, info: {show_slider: true, is_inner: false}, permission: 'system.menu'},
    },
    {
        name: 'systemConfigNode',
        path: '/system/config/node',
        component: resolve => require(['@/views/system/config/node'], resolve),
        meta: {model: 40, info: {show_slider: true, is_inner: false}, permission: 'system.node'},
    },
];
