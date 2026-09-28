<script lang="ts">
	import { cn } from '$lib/utils.js';
	let klass: string | undefined = undefined;
	export { klass as class };
</script>

<div class={cn('shrink-0 bg-border h-[1px] w-full', klass)} {...$$restProps} />
