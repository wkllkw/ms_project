<template>
    <div class="page-404">
        <ErrorPage
                code="404"
                desc="抱歉，你访问的页面不存在"
                subDesc="请检查输入的网址是否正确，或返回首页浏览其他内容"
        >
            <img slot="img" src="../../assets/image/error/404.svg" alt="">
        </ErrorPage>
    </div>
</template>
<script>
    import ErrorPage from '../../components/error/errorPage';
    export default {
        components: {
            ErrorPage
        }
    }
</script>
